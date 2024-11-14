DO $$
    BEGIN
        IF EXISTS (
            SELECT 1
            FROM pg_constraint
            WHERE conname = 'users_email_unique'
        ) THEN
    ALTER TABLE users
    DROP CONSTRAINT users_email_unique;
    END IF;
END $$;
