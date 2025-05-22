# Keys & Authentication

## Generate Hashed Key

```bash
go run main.go -key yourpassword
```

Use the output hash in your `keys.json` for authentication.

## Keys.JSON format
Create a `keys.json` file in your (webserver) keys directory.

Example (default [`file`](../keys.json) from Mazarin):
```json
{
  "users": [
      {
        "name": "test",
        "hash": "$2a$10$f.qQVxQMikTkKZWYekqYfOi17O8f1/83HA5CX8TADYtQGhHmptZha",
      },
      {
        "name": "user2",
        "hash": "$2a$10$Z1/wTrjFwzWaC60CwQYgVe.M7hcKr0YESo2G6etOSInxkklltcfIO", 
      }
  ]
}
```
(In this example the password for test is test_password and for user2 is user2_password)

- **name:** The username of the user.
- **hash:** The generated hash of `go run main.go -key yourpassword`