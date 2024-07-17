CREATE TABLE IF NOT EXISTS permissions (
    id BIGSERIAL PRIMARY KEY,
    code text NOT NULL
);

CREATE TABLE IF NOT EXISTS users_permissions(
    user_id BIGINT NOT NULL REFERENCES users ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions on DELETE CASCADE,
    PRIMARY KEY (user_id, permission_id)
    -- The PRIMARY KEY (user_id, permission_id) line sets a composite primary key on our users_permissions table,the
    -- primary key is made up of both the users_id and permission_id columns. Setting this as the primary key means
    -- that the same user/permission combination can only appear once in the table and cannot be duplicated.
);

INSERT INTO permissions (code)
VALUES
('movies:read'),
('movies:write');