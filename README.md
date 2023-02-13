# Movie Sunday User Service

### Curl Commands
For my Alzheimer's brain

```
curl -i -k -X POST -H "Authorization: <token>" -H "Username: dan" https://localhost:4321/user -H 'Content-Type: application/json' -d '{"username":"dan","password":"password","email":"email@email.com","created_on":"2015-07-01"}'

curl -i -k -X GET -H "Authorization: <token>" -H "Username: dan" https://localhost:4321/users

curl -i -k -X GET -H "Authorization: <token>" -H "Username: dan" https://localhost:4321/user/dan

curl -i -k -X GET -H "Authorization: Bearer <token>" -H "Username: dan" https://localhost:4321/validate

curl -i -k -X POST https://localhost:4321/login -u "dan:password"
```

### Create SSL Keys
```
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./secrets/key.key -out ./secrets/cert.crt 
```

### Docker

```
docker build -t ms_user_service .
docker run -p 4321:4321 -tid ms_user_service
```
