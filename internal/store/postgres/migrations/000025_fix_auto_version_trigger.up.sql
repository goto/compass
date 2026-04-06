CREATE
OR REPLACE FUNCTION auto_version_trigger()
    RETURNS TRIGGER AS $$
BEGIN
    IF
(TG_OP = 'INSERT') THEN
        NEW.version := '0.1';
END IF;

RETURN NEW;
END;
$$
LANGUAGE plpgsql;
