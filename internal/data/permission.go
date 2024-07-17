package data

import "github.com/jackc/pgx/v5/pgxpool"

type Permissions []string

func (p Permissions) Include(code string) bool {
	for i := range p {
		if code == p[i] {
			return true
		}
	}
	return false
}

type PermissionModel struct {
	DB *pgxpool.Pool
}

func (m PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	query := `
		SELECT p.code 
		FROM permissions p
		INNER JOIN users_permissions up ON up.permission_id = p.id
		INNER JOIN users u ON u.id = up.user_id
		WHERE u.id = $1
		`
	ctx, cancel := newQueryContext(3)
	defer cancel()
	rows, _ := m.DB.Query(ctx, query, userID)
	defer rows.Close()
	var permissions Permissions
	for rows.Next() {
		var permission string
		if err := rows.Scan(&permission); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return permissions, nil
}

func (m PermissionModel) AddForUser(userID int64, codes ...string) error {
	query := `
		INSERT INTO users_permissions
		SELECT $1, p.id FROM permissions p WHERE p.code = ANY($2)
		`
	ctx, cancel := newQueryContext(3)
	defer cancel()
	_, err := m.DB.Exec(ctx, query, userID, codes)
	return err
}
