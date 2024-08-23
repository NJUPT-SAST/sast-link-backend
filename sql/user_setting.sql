CREATE TABLE public.user_setting (
    user_id VARCHAR(255) NOT NULL,       -- 'user_id' is the user identifier
    key INT NOT NULL,                    -- 'setting_key' represents the setting type
    value TEXT NOT NULL,                 -- 'value' stores the JSON string or other setting values
    PRIMARY KEY (user_id, key)           -- Combination of 'user_id' and 'key' as the primary key
);
