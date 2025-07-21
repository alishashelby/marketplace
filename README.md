# marketplace

## link:
### get all ads:
http://185.207.0.69:8081/api/ads?page=1&limit=10&sort_by=created_at&order_by=1&min_price=100&max_price=10000.5

### register:
http://185.207.0.69:8081/api/register

body:
```cpp
{
"username": "john",
"password": "12345&678"
}
```

### login:
http://185.207.0.69:8081/api/login

body:
```cpp
{
"username": "john",
"password": "12345&678"
}
```

### get all ads with owner marker:
http://185.207.0.69:8081/api/ads/?page=1 - and other query params

### publish ad:
http://185.207.0.69:8081/api/publish

body:
```cpp
{
"title": "title of test ad",
"text": "this is the test ad. check new image.",
"image_url": "https://upload.wikimedia.org/wikipedia/commons/c/c7/Tabby_cat_with_blue_eyes-3336579.jpg",
"price": 1500.5
}
```