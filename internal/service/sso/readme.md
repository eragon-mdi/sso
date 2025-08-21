## Register

Что делает: создаёт учётную запись.
Вход: User{email, password}.
Выход: user_id (UUID).
Что происходит (сервер):

Валидация входа (формат email, длина пароля).

Хеширование пароля (bcrypt/scrypt) → создание User (генерация id).

Сохранение в UserStore.
gRPC статусы:

OK — создан.

AlreadyExists — email занят.

InvalidArgument — неверные данные.

Internal — ошибка сервиса/БД.
Для пользователя: регистрируется и получает id; не получает токены — для этого отдельный Login.

## Login

Что делает: аутентифицирует по паролю и выдаёт пару токенов (access + refresh).
Вход: User{email,password}, DeviceContext{app_id,device_id}.
Выход: TokenPair{access, refresh}.
Что происходит (сервер):

Валидация входа.

Достаёт запись пользователя из UserStore (hash пароля).

Сравнивает пароли (passHasher.Compare). Если ок — продолжает.

Формирует Meta (UserID + ctx + exp).

Tokener.GenPair(meta) → получает access и refresh (строки).

Хэширует refresh (TokenHasher.Sum) и сохраняет запись в refresh_tokens с meta (user_id, app_id, device_id, jti, exp, revoked=false).
gRPC статусы:

OK — выданы токены.

Unauthenticated — неверный пароль.

InvalidArgument — неверный запрос.

Internal — ошибка при генерации/сохранении.
Для пользователя: вводит логин/пароль → получает короткоживущий access (для запросов) и долгоживущий refresh (для обновления).

## Refresh

Что делает: по refresh токену возвращает новую пару (ротация).
Вход: refresh (строка), DeviceContext.
Выход: TokenPair{access, refresh} (новая пара).
Что происходит (сервер):

Верификация подписи и exp токена через Tokener.VerifyRefresh.

Сравнение claims из токена с пришедшим DeviceContext (app_id/device_id).

Хэширование пришедшего refresh → поиск в refresh_tokens по hash + user_id + app_id + device_id.

Если запись не найдена / revoked / expired → Unauthenticated.

Если валидно → генерируем новую пару (GenPair), считаем hash(newRefresh).

Выполняем атомарную ротацию: Rotate(oldHash, newRefreshRecord) (revoke old + insert new в транзакции).

Возвращаем новые токены.
gRPC статусы:

OK — новые токены.

Unauthenticated — невалидный/просроченный/отозванный токен или ctx mismatch.

Internal — ошибка БД/ротации.
Для пользователя: клиент отправляет refresh → получает новый access и новый refresh; если refresh скомпрометирован, сервис откатывает сессии и требует повторный логин.

## Logout

Что делает: отзывает конкретный refresh (идемпотентно).
Вход: refresh (строка), DeviceContext.
Выход: пустой (google.protobuf.Empty).
Что происходит (сервер):

Верификация подписи/exp через Tokener.VerifyRefresh.

Сравнение claims с DeviceContext.

Хэширование refresh → атомарный UPDATE ... SET revoked = TRUE WHERE hash = $1 AND user_id = $2 AND app_id = $3 AND device_id = $4 AND revoked = FALSE.

Если rows_affected==1 → OK. Если 0 → уже отозвано/не найдено → тоже OK (идемпотентно).
gRPC статусы:

OK — успешно/идемпотентно.

Unauthenticated — невалидный токен или ctx mismatch.

Internal — ошибка БД.
Для пользователя: нажал «выйти» — сессия на данном устройстве/приложении отозвана.