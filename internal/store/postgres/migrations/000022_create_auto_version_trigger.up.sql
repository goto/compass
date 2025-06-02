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
BEGIN
    -- Check if any field other than refreshed_at has changed
    IF (
        NEW IS DISTINCT FROM OLD AND
        (
            (NEW.* IS DISTINCT FROM OLD.*) AND
            NOT (NEW.refreshed_at IS DISTINCT FROM OLD.refreshed_at)
            )
        ) THEN
        NEW.version := bump_minor_version(OLD.version);
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