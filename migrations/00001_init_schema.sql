-- +goose Up

CREATE TABLE departments (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    parent_id BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,

    CONSTRAINT fk_department_parent FOREIGN KEY (parent_id)
        REFERENCES departments(id) ON DELETE RESTRICT

    CONSTRAINT uq_department_name_parent UNIQUE (parent_id, name)
);

CREATE INDEX idx_department_parent_id ON departments(parent_id);

CREATE UNIQUE INDEX uq_departments_root_name ON departments(name) WHERE parent_id IS NULL;

CREATE TABLE employees (
    id BIGSERIAL PRIMARY KEY,
    department_id BIGINT NOT NULL,
    full_name VARCHAR(200) NOT NULL,
    position VARCHAR(200) NOT NULL,
    hired_at DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,

    CONSTRAINT fk_employee_department FOREIGN KEY (department_id)
        REFERENCES departments(id) ON DELETE CASCADE
);

CREATE INDEX idx_employee_department_id ON eployees(department_id);

-- +goose Down
DROP TABLE IF EXIST empoyees;
DROP TABLE IF EXIST departments;
