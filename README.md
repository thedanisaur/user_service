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
TLS keys
```
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./secrets/key.key -out ./secrets/cert.crt 
sudo chmod 775 ./secrets/cert.crt 
sudo chmod 775 ./secrets/key.key
```

http keys
```
sudo openssl genrsa -out ./secrets/key.key 3072
sudo openssl rsa -in ./secrets/key.key -pubout -out ./secrets/cert.crt
sudo openssl pkcs8 -topk8 -inform PEM -outform PEM -nocrypt -in ./secrets/key.key -out ./secrets/key8.key
```

### Docker

```
docker build -t ms_user_service .
docker run -p 4321:4321 -tid ms_user_service
```

### Service
setup systemd service
```
sudo cp user_service.service /lib/systemd/system/.
sudo chmod 755 /lib/systemd/system/user_service.service
sudo systemctl daemon-reload
sudo systemctl enable user_service.service
sudo systemctl start user_service
```

view logs
```
sudo journalctl -u user_service
```

remove service
```
sudo systemctl stop user_service.service
sudo systemctl disable user_service.service
sudo rm /etc/systemd/system/user_service.service
sudo rm /usr/lib/systemd/system/user_service.service
sudo systemctl daemon-reload
sudo systemctl reset-failed
```