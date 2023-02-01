# user_auth

### Curl Commands
For my Alzheimer's brain

```
curl -i -X POST -H "Authorization: <token>" -H "Username: dan" http://localhost:4321/user -H 'Content-Type: application/json' -d '{"username":"dan","password":"password","email":"definisdan@yahoo.com","created_on":"2015-07-01"}'

curl -i -X GET -H "Authorization: <token>" -H "Username: dan" http://localhost:4321/users

curl -i -X GET -H "Authorization: <token>" -H "Username: dan" http://localhost:4321/user/dan

curl -i -X POST http://localhost:4321/login -u "dan:password"
```
