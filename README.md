# api.edwica
Подбор видео, книг, вакансий для навыков


## Деплой

Суть этого деплоя можно выразить в следующих шагах:
1. Пишем какое-то веб-приложение (сервер, сайт, апи)
2. Кидаем его на сервер
3. Покупаем для него доменное имя
4. Настраиваем для него nginx-конфиг
5. Настраиваем для него systemd-конфиг


### Шаг первый: Написать приложение
Рассматриваться деплой приложения будет на примере REST-API, в котором доступны только GET-методы. Методы и функциональность не имеют никакого отношения к деплою
Наше API работает с БД и парсить HTML-страницы. Выдает список книг, видео и вакансий по запросу

### Шаг второй: Загрузка на сервер
Приложение написали, запушили его на гитхаб. Теперь подключаемся к нашему серверу и клонируем туда проект. Настраиваем переменные окружения и подтягиваем зависимости. В этот раз пришлось еще обновить версию Go.

Как обновить Go, если он у вас уже установлен?
```bash
sudo apt-get purge golang*
wget https://go.dev/dl/go1.21.4.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.21.4.linux-amd64.tar.gz
go version
```

### Шаг третий: Покупка домена
Домен за нас уже купили и даже подключили его к нашему серверу. Поэтому мы можем смело разворачивать на нем свои приложения
Доменное имя в данном случае такое: **api.edwica.ru**
Так как к нему уже подключен FastAPI проект на Python 81 портом, то мы можем подключить еще одно приложение заменив порт на другой

### Шаг четвертый: Настройка nginx
Установка:
```bash
sudo apt install nginx
sudo systemctl start nginx
sudo systemctl enable nginx
```
Создаем новый конфиг для нашего приложения. Называем его в данном случае **go-edwica-api.ru** Точка обязательна
```bash
nano /etc/nginx/sites-available/go-edwica-api.ru
```
Пишем в файл go-edwica-api.ru следующее:
```shell
server {
    listen 88; # Порт, по которому будет доступен сайт: api.edwica.ru:88
    server_name api.edwica.ru; # доменное имя
    location / {
        proxy_pass http://127.0.0.1:8989; # порт, к которому подключается приложение
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```
*Копируем* файл в другую директорию
```shell
sudo ln -s /etc/nginx/sites-available/go-edwica-api.ru /etc/nginx/sites-enabled/
sudo systemctl restart nginx
```
Настраиваем SSL-сертификат
```shell
sudo apt install certbot python3-certbot-nginx
certbot --nginx -d 'go-edwica-api.ru'
sudo systemctl restart nginx
```

### Шаг пятый: Настройка systemctl
Этот этап позволяет запускать наше приложение вместе с сервером. Например, если сервер сдохнет на время и оживет спустя несколько часов, наше приложение автоматически поднимется за ним. 

Создаем конфиг-файл:
```shell
nano /etc/systemd/system/go-edwica-api.service
```
Пишем туда:
```shell

[Unit]
    Description=Golang API api.edwica.ru # Заменить на свое
    After=network.target

[Service]
    User=root
    Group=root

    WorkingDirectory=/root/go/src/github.com/Rosya-edwica/api.edwica/ # Заменить на свое
    ExecStart=/root/go/src/github.com/Rosya-edwica/api.edwica/api # Заменить на свое
[Install]
        WantedBy=multi-user.target
```
Перезапускаем систему:
```shell
systemctl enable go-edwica-api.service
systemctl start go-edwica-api.service
systemctl status go-edwica-api.service
```

### Итог:
Приложение запущено и работает успешно! Если нужно поменять что-то в приложении и применить изменения:

1.  Заново компилируем приложение
```shell
go build -o api
```
2.  Перезапускаем демона
```shell
systemctl daemon-reload
```
3. Перезапускаем приложение
```shell
systemctl restart go-edwica-api.service
```
