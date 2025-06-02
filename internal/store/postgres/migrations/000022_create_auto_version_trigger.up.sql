-- 1. Function to bump the minor version
CREATE OR REPLACE FUNCTION bump_minor_version(current_version TEXT)
    RETURNS TEXT AS $$
DECLARE
    major INTEGER;
    minor INTEGER;
BEGIN
    IF current_version IS NULL OR current_version !~ '^\d+\.\d+$' THEN
        RETURN '0.1';
    END IF;

    major := split_part(current_version, '.', 1)::INTEGER;
    minor := split_part(current_version, '.', 2)::INTEGER + 1;

    RETURN major || '.' || minor;
END;
$$ LANGUAGE plpgsql;

-- 2. Trigger function to auto-bump version
CREATE OR REPLACE FUNCTION auto_version_trigger()
    RETURNS TRIGGER AS $$
DECLARE
    changed_columns TEXT[];
BEGIN
    -- For INSERT operations, set initial version to "0.1"
    IF (TG_OP = 'INSERT') THEN
        NEW.version := '0.1';

    -- For UPDATE operations
    ELSIF (TG_OP = 'UPDATE') THEN
        -- Get list of changed columns (excluding refreshed_at)
        SELECT array_agg(key) INTO changed_columns
        FROM jsonb_each(to_jsonb(NEW)) AS n
                 JOIN jsonb_each(to_jsonb(OLD)) AS o USING (key)
        WHERE n.key != 'refreshed_at' AND n.value IS DISTINCT FROM o.value;

        -- If any columns other than refreshed_at changed, bump version
        IF changed_columns IS NOT NULL AND array_length(changed_columns, 1) > 0 THEN
            NEW.version := bump_minor_version(OLD.version);
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 3. Create trigger on updates
DROP TRIGGER IF EXISTS update_version_trigger ON assets;
CREATE TRIGGER update_version_trigger
    BEFORE UPDATE ON assets
    FOR EACH ROW
EXECUTE FUNCTION auto_version_trigger();