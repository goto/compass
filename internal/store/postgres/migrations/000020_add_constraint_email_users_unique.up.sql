DO $$
    BEGIN
        IF NOT EXISTS (
            SELECT 1
            FROM pg_constraint
            WHERE conname = 'users_email_unique'
        ) THEN
    ALTER TABLE users
        ADD CONSTRAINT users_email_unique UNIQUE (email);
    END IF;
END $$;
