CREATE TABLE public.system_setting (
    name VARCHAR(255) PRIMARY KEY,   -- 'name' is a unique identifier
    value TEXT NOT NULL,             -- Storing the 'value' as text, as it could vary in length
    description TEXT                 -- 'description' is optional, hence no NOT NULL constraint
);
