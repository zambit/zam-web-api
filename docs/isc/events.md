# User event

This section describes messages which is sent when some key events occurs to an specific user.

All events emitted from resource's with name `users`.

## Registration events

Events which occurs during user registration.

### **EVENT:** `users.registration_verification_required_event.{user_id}`

Emitted when user phone registration is required during registration process

Params:

1) `user_id`
    * Type: string
    * Description: new user identifier

2) `user_phone`
    * Type: string
    * Format: phone_number
    * Description: user phone

3) `verification_code`
    * Type: string
    * Description: verification code which should be sent by user on next `../signup/verify` request

### **EVENT:** `user.registration_verification_completed_event.{user_id}`

Emitted when user completes registration process

Params:

1) `user_id`
    * Type: string
    * Description: affected user identifier
    
2) `user_phone`
    * Type: string
    * Format: phone_number
    * Description: user phone

## Password recovery events

Events which occurs during user password recovery process.

### **EVENT:** `users.password_recovery_verification_required_event.{user_id}`

Emitted when user should verify password recovery.

Params:

1) `user_id`
    * Type: string
    * Description: affected user identifier

2) `user_phone`
    * Type: string
    * Format: phone_number
    * Description: user phone

3) `recovery_code`
    * Type: string
    * Description: verification code which should be sent by user on next `../recovery/verify` request

### **EVENT:** `user.password_recovery_completed_event.{user_id}`

Emitted when user completes password recovery

Params:

1) `user_id`
    * Type: string
    * Description: affected user identifier

2) `user_phone`
    * Type: string
    * Format: phone_number
    * Description: user phone
