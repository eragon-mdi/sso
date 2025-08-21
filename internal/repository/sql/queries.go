package sqlrepo

// --- AUTH ---
const queryInsertUser = `
INSERT INTO users (id, email, password_hash)
VALUES ($1, $2, $3)
RETURNING id, email, password_hash
`

const queryGetUserByEmail = `
SELECT id, email, password_hash
FROM users
WHERE email = $1
`

// --- PERMISSION ---
const queryCheckUserIsAdmin = `
SELECT EXISTS (
	SELECT 1 FROM user_roles WHERE user_id = $1 AND role = 'admin'
)
`
