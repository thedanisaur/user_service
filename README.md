# user_auth

### Curl Commands
For my Alzheimer's brain

```
curl -i -k -X POST -H "Authorization: <token>" -H "Username: dan" https://localhost:4321/user -H 'Content-Type: application/json' -d '{"username":"dan","password":"password","email":"definisdan@yahoo.com","created_on":"2015-07-01"}'

curl -i -k -X GET -H "Authorization: <token>" -H "Username: dan" https://localhost:4321/users

curl -i -k -X GET -H "Authorization: <token>" -H "Username: dan" https://localhost:4321/user/dan

curl -i -k -X POST https://localhost:4321/login -u "dan:password"
```

### Create SSL Keys
```
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./keys/key.key -out ./certs/cert.crt 
```
