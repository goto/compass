-- Drop the trigger on the assets table
DROP TRIGGER IF EXISTS update_version_trigger ON assets;

-- Drop the trigger function
DROP FUNCTION IF EXISTS auto_version_trigger;

-- Drop the version bump logic function
DROP FUNCTION IF EXISTS bump_minor_version;