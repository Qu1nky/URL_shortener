Требуется создать микросервис сокращения url. Длина сокращенного URL-адреса должна быть как
можно короче. Сокращенный URL может содержать цифры (0-9) и буквы (a-z, A-Z)
Эндпоинты:
POST http://localhost:8080/
Request: (body): http://cjdr17afeihmk.biz/123/kdni9/z9d112423421
Response: http://localhost:8080/qtj5opu
GET
Request (url query): http://localhost:8080/qtj5opu
Response (body): http://cjdr17afeihmk.biz/123/kdni9/z9d112423421
Микросервис должен уметь хранить информацию в памяти и в postgres в зависимости от флага
запуска -d
