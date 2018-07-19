package jwt

import (
	"fmt"
	"github.com/google/uuid"
	"gitlab.com/ZamzamTech/wallet-api/services/nosql"
	"gitlab.com/ZamzamTech/wallet-api/services/sessions"
	"gopkg.in/dgrijalva/jwt-go.v3"
	"time"
)

const (
	tokenPersisIDKey = "persistKey"
	expireAtKey      = "exp"
	issuedAtKey      = "iat"
)

// jwtStorage implements storage interface using jwt-token mechanism where all optional data stored on the user side.
// If persistent storage are passed, it also allow to track deleted sessions.
type jwtStorage struct {
	nowFunc func() time.Time

	signingMethod jwt.SigningMethod
	secretKey     []byte

	storageKeyFunc    func(data map[string]interface{}, token string) string
	persistentStorage nosql.IStorage
}

// New creates new jwt storage without ability to delete token, panic on wrong signing method
func New(signingMethod string, secret []byte, nowFunc func() time.Time) sessions.IStorage {
	switch signingMethod {
	case "RS256", "RS512", "RS384":
		panic(fmt.Errorf(
			`public-key signing alghoritms not supported ("RS256", "RS512", "RS384"): given value is %s`, signingMethod,
		))
	}

	return &jwtStorage{
		nowFunc:       nowFunc,
		secretKey:     secret,
		signingMethod: jwt.GetSigningMethod(signingMethod),
	}
}

// WithStorage jwt storage which uses persistent storage
func WithStorage(
	storage sessions.IStorage,
	persistentStorage nosql.IStorage,
	storageKeyFunc func(data map[string]interface{}, token string) string,
) sessions.IStorage {
	jwtSt, ok := storage.(*jwtStorage)
	if !ok {
		panic(fmt.Sprintf("expect %T, but %T received", &jwtStorage{}, jwtSt))
	}

	jwtSt.storageKeyFunc = storageKeyFunc
	jwtSt.persistentStorage = persistentStorage
	return jwtSt
}

// New generates new jwt token and track it if persistent storage is present
func (s *jwtStorage) New(data map[string]interface{}, expireAfter time.Duration) (sessions.Token, error) {
	token := jwt.New(s.signingMethod)
	claims := token.Claims.(jwt.MapClaims)

	// copy data
	for key, val := range data {
		claims[key] = val
	}

	// populate
	claims[issuedAtKey] = s.nowFunc().Unix()
	claims[expireAtKey] = s.nowFunc().Add(expireAfter).Unix()

	// if token storage is defined, use it to store token session
	// it's allow us to track deleted sessions
	if s.persistentStorage != nil {
		tokenID := uuid.New().String()
		claims[tokenPersisIDKey] = tokenID
		err := s.persistentStorage.SetWithExpire(s.storageKeyFunc(data, tokenID), tokenID, expireAfter)
		if err != nil {
			return sessions.Token{}, err
		}
	}

	// generate token string
	tokenString, err := token.SignedString(s.secretKey)
	return sessions.Token(tokenString), err
}

// Get validates token, checks expiration and checks persistent storage for such token if it present
func (s *jwtStorage) Get(token sessions.Token) (data map[string]interface{}, err error) {
	claims, err := s.extractClaimsFromToken(token)
	if err != nil {
		return
	}

	// check token expired
	expireAtTs, err := extractTimestampFromClaims(claims, expireAtKey)
	if err != nil {
		return
	}
	if s.nowFunc().Unix() > expireAtTs {
		err = sessions.ErrExpired
		return
	}

	// if token storage defined, check token
	if s.persistentStorage != nil {
		tokenID, err := extractTokenIDFRomClaims(claims)
		if err != nil {
			return nil, err
		}

		// lookup token
		if _, err = s.persistentStorage.Get(s.storageKeyFunc(claims, tokenID)); err != nil {
			if err == nosql.ErrNoSuchKeyFound {
				err = sessions.ErrNotFound
			}
		}
		if err != nil {
			return nil, err
		}
	}

	// copy payload values except technical one
	data = make(map[string]interface{}, len(claims))
	for key, val := range claims {
		if key == expireAtKey || key == issuedAtKey || key == tokenPersisIDKey {
			continue
		}
		data[key] = val
	}

	return
}

// RefreshToken
func (s *jwtStorage) RefreshToken(oldToken sessions.Token, expireAfter time.Duration) (sessions.Token, error) {
	data, err := s.Get(oldToken)
	if err != nil {
		return sessions.Token{}, err
	}
	return s.New(data, expireAfter)
}

// Delete token associated id in persistent storage id it present
func (s *jwtStorage) Delete(token sessions.Token) error {
	// if there is no session storage, jwt token can't be deleted
	if s.persistentStorage == nil {
		// maybe there should be some kind of warning?
		return nil
	}

	// extract claims
	claims, err := s.extractClaimsFromToken(token)
	if err != nil {
		return err
	}

	// extract token id
	tokenID, err := extractTokenIDFRomClaims(claims)
	if err != nil {
		return err
	}

	// remove it from tokens storage
	err = s.persistentStorage.Delete(s.storageKeyFunc(claims, tokenID))
	if err == nosql.ErrNoSuchKeyFound {
		return sessions.ErrNotFound
	}
	return err
}

func (s *jwtStorage) extractClaimsFromToken(token sessions.Token) (jwt.MapClaims, error) {
	decodedToken, err := jwt.Parse(string(token), func(token *jwt.Token) (interface{}, error) {
		if token.Method != s.signingMethod {
			return nil, fmt.Errorf("%s: invalid signing method: %s", sessions.ErrUnexpectedToken.Error(), token.Method)
		}
		return s.secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	return decodedToken.Claims.(jwt.MapClaims), err
}

// utils
func extractTimestampFromClaims(claims jwt.MapClaims, key string) (int64, error) {
	if expireAt, ok := claims[key]; ok {
		if expireAt, ok := expireAt.(float64); ok {
			return int64(expireAt), nil
		}
	}
	return 0, sessions.ErrUnexpectedToken
}

func extractTokenIDFRomClaims(claims jwt.MapClaims) (tokenID string, err error) {
	// validate payload token id
	if tokenIDRaw, ok := claims[tokenPersisIDKey]; ok {
		tokenID, ok = tokenIDRaw.(string)
		if !ok {
			err = sessions.ErrUnexpectedToken
		}
	} else {
		err = sessions.ErrUnexpectedToken
	}
	return
}
