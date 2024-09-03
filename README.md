# go-auth-test-task

Для запуска используется docker compose:
docker compose up

Пароль и имя пользователя базы данных конфигурируются в файле .env вместе с секретами.

Для отправки писем используется smtp сервер google, для корректной работы которого с plain auth методом авторизации необоходимо сначала сделать пароль для приложения (https://security.google.com/settings/security/apppasswords) и затем записать её значение в env переменную EMAIL_PASSWORD. 

Также, необходимо указать адрес почты отправителя в переменной EMAIL_ADDRESS.

Работа проверена на Arch Linux с Linux 6.10.0-arch1-2.
