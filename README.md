# BH_Lu_3

## Описание

Распределённый вычислитель арифметических выражений с реализованной афторизацие
пользователей. Представляет из себя три микро сервера

1: BH - отвечает за афторизацию, выдачю JWT и взаимодейтвия с базой данных, а так-же
управляет работой всего проекта. Запущен на порту 9090.
2: Orchestrator - отвечает за дробления выражений на Tasks для последующего вычисления. Запущен на порту 8080.
3: Lu - отвечает за вычисление Task и отправку результата orchestratoru. Запущен на порту 8081.

## Запуск

•Установите [Docker](https://www.docker.com/).
**Приложение Docker Dector должно быть запущено и коректно работать на момент развёртывания BH_Lu_3**.

•Откройте cmd от имени администратора и cклонируйте проект с GitHub.

```bash
 git clone https://github.com/nikitakutergin59/BH_Lu_3
```

•Перейдите в папку с проектом.

```bash
 cd ./BH_Lu_3
```

•Запустите и скомпилируйте docker-compose.

```bash
 docker-compose up --build
```

-**Вы должыны увидеть примерно вот это**:

   [+] Running 7/7

 ✔ bh                              Built                                                                              0.0s
 ✔ demon                           Built                                                                              0.0s
 ✔ orchestrator                    Built                                                                              0.0s
 ✔ Network bh_lu_3_default         Created                                                                            0.0s
 ✔ Container orchestrator_service  Created                                                                            0.3s
 ✔ Container bh_service            Created                                                                            0.3s
 ✔ Container demon_service         Created                                                                            0.3s
 Attaching to bh_service, demon_service, orchestrator_service
 orchestrator_service  | 2025/05/09 21:39:24 Оркестратор запущен на порту 8080
 demon_service         | 2025/05/09 21:39:24 Демон запущен на порту 8081
 bh_service            | 2025/05/09 21:39:24 Таблица user_data успешно создана или уже существует.
 bh_service            | 2025/05/09 21:39:24 База данных созданна
 bh_service            | 2025/05/09 21:39:24 Сервер запущен на порту 9090

Не забывайте смотреть логи, там очень много важной информации.

### Возможные проблемы

•**Не включена виртуализация в BIOS.**
•**Не скачан Linux для Windows(обычно должен скачиваться вместе с Docker)**
•**Порт 9090, 8080 или 8081 уже занят другим приложением.**

### Взаимодействия и примеры запросов

Откройте cmd (Win+R)

#### 1. Регистраци

 **Отправте запрос(придумайте логин и пароль)**:

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"login":"your_login", "password":"your_password"}' \
  http://localhost:9090/register
```

 -**Вы получите ответ, пример**:

#### 2. Авторизация

 **Отправте запрос**:

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"login":"your_hash_login", "password":"your_hash_password"}' \
  http://localhost:9090/login
```

 -**Например**:

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"login":"dda844144da6949024dc7cc621a47f73d0b3357fd1f5c5c4f9167bf026414aa7", "password\":"b5b6f2b7707de42254c0e13ce2a2f53ce9e4bbbd282c35cb61ba70470c331440"}' \
  http://localhost:9090/login
```

 -**Вы получите ответ пример**:

```bash
{"message":"успешный вход","token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9eyJzdWIiOiJkZGE4NDQxNDRkYTY5NDkwMjRkYzdjYzYyMWE0N2Y3M2QwYjMzNTdmZDFmNWM1YzRmOTE2N2JmMDI2NDE0YWE3IiwiZXhwIjoxNzQ2NzI5NzIwLCJpYXQiOjE3NDY2NDMzMjB9.m3mwt3mgpcnoTQ23cEeVZUkyDP5eZEhA03jqiMgwAY0"}
```

 //Да уж он получился очень длинный(это нормально)

 -**Важно**:
 •Токен действует только 24 часа!

 •Если вы столкнулись с тем что вы всё ввели правильно, а BH_Lu_3 выдаёт ошибку **неверный логин или пароль**
 то придумайте новый логин и пароль и попробуйте занова(смотреть пункт 1)

#### 3. Вычисления

 Поздравляю вы афторизировались и теперь можно пользоваться всем доступный функционалом.

 -**Создайте запрос на вычисление**:

```bash
curl -X POST \
  -H "Authorization: Bearer <your_jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{"expr": "<your_expression>"}' \
  http://localhost:9090/calculator
```

 -**Пример**:

```bash
curl -X POST \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJkZGE4NDQxNDRkYTY5NDkwMjRkYzdjYzYyMWE0N2Y3M2QwYjMzNTdmZDFmNWM1YzRmOTE2N2JmMDI2NDE0YWE3IiwiZXhwIjoxNzQ2NzI5NzIwLCJpYXQiOjE3NDY2NDMzMjB9.m3mwt3mgpcnoTQ23cEeVZUkyDP5eZEhA03jqiMgwAY0" \
  -H "Content-Type: application/json" \
  -d '{"expr": "2*(5+3)-(4+6)/2"}' \
  http://localhost:9090/calculator
```

 -**В ответ вы получите, пример**:

```bash
{"message":"выражение обработано","result":"11","token_id":"dda844144da6949024dc7cc621a47f73d0b3357fd1f5c5c4f9167bf026414aa7"}
```

 -**Вот ещё несколько примеров**:

```bash
curl -X POST \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJkZGE4NDQxNDRkYTY5NDkwMjRkYzdjYzYyMWE0N2Y3M2QwYjMzNTdmZDFmNWM1YzRmOTE2N2JmMDI2NDE0YWE3IiwiZXhwIjoxNzQ2NzI5NzIwLCJpYXQiOjE3NDY2NDMzMjB9.m3mwt3mgpcnoTQ23cEeVZUkyDP5eZEhA03jqiMgwAY0" \
  -H "Content-Type: application/json" \
  -d '{"expr": "3+(2*(7-4))"}' \
  http://localhost:9090/calculator
```

 -**Ответ**:

```bash
{"message":"выражение обработано","result":"1","token_id":"dda844144da6949024dc7cc621a47f73d0b3357fd1f5c5c4f9167bf026414aa7"}
```

 -**Пример**:

```bash
curl -X POST \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJkZGE4NDQxNDRkYTY5NDkwMjRkYzdjYzYyMWE0N2Y3M2QwYjMzNTdmZDFmNWM1YzRmOTE2N2JmMDI2NDE0YWE3IiwiZXhwIjoxNzQ2NzI5NzIwLCJpYXQiOjE3NDY2NDMzMjB9.m3mwt3mgpcnoTQ23cEeVZUkyDP5eZEhA03jqiMgwAY0" \
  -H "Content-Type: application/json" \
  -d '{"expr": "(10+5)*(6-2)"}' \
  http://localhost:9090/calculator
```

 -**Ответ**:

```bash
{"message":"выражение обработано","result":"60","token_id":"dda844144da6949024dc7cc621a47f73d0b3357fd1f5c5c4f9167bf026414aa7"}
```

 -**Пример**:

```bash
curl -X POST \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJkZGE4NDQxNDRkYTY5NDkwMjRkYzdjYzYyMWE0N2Y3M2QwYjMzNTdmZDFmNWM1YzRmOTE2N2JmMDI2NDE0YWE3IiwiZXhwIjoxNzQ2NzI5NzIwLCJpYXQiOjE3NDY2NDMzMjB9.m3mwt3mgpcnoTQ23cEeVZUkyDP5eZEhA03jqiMgwAY0" \
  -H "Content-Type: application/json" \
  -d '{"expr": "((8+2)*3-4)/2"}' \
  http://localhost:9090/calculator
```

 -**Ответ**:

```bash
{"message":"выражение обработано","result":"2","token_id":"dda844144da6949024dc7cc621a47f73d0b3357fd1f5c5c4f9167bf026414aa7"}
```

 **Если вы введёте выражение которое не может быть посчитано или введёте нечего то поле result будет пустым, но ошибка обработанна и информация об этом храниться в логах**:

#### 4. Получение выражений и результатов по JWT

Отправте запрос

```bash
curl -X GET \
  -H "Authorization: Bearer <your_jwt_token>" \
  http://localhost:9090/my_data
```

-**Пример**:

```bash
curl -X GET \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9eyJzdWIiOiJkZGE4NDQxNDRkYTY5NDkwMjRkYzdjYzYyMWE0N2Y3M2QwYjMzNTdmZDFmNWM1YzRmOTE2N2JmMDI2NDE0YWE3IiwiZXhwIjoxNzQ2NzI5NzIwLCJpYXQiOjE3NDY2NDMzMjB9.m3mwt3mgpcnoTQ23cEeVZUkyDP5eZEhA03jqiMgwAY0" \
  http://localhost:9090/my_data
```

-**Ответ**:

```bash
{"expressions":"2*(5+3)-(4+6)/2;3+(2*(7-4));(10+5)*(6-2);;10/0;((8+2)*3-4)/2;","results":"11;1;60;;;2;"}
```

## Заключение

Для завершения работы комбинация клавишь Ctrl+C затем Enter затем введите команду.

```bash
docker-compose down
```

BH
