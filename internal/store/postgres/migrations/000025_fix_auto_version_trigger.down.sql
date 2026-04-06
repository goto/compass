CREATE
OR REPLACE FUNCTION auto_version_trigger()
    RETURNS TRIGGER AS $$
DECLARE
changed_columns TEXT[];
BEGIN
    IF
(TG_OP = 'INSERT') THEN
        NEW.version := '0.1';

    ELSIF
(TG_OP = 'UPDATE') THEN
SELECT array_agg(key)
INTO changed_columns
FROM jsonb_each(to_jsonb(NEW)) AS n
         JOIN jsonb_each(to_jsonb(OLD)) AS o USING (key)
WHERE n.key != 'refreshed_at' AND n.value IS DISTINCT
FROM o.value;

IF
changed_columns IS NOT NULL AND array_length(changed_columns, 1) > 0 THEN
            NEW.version := bump_minor_version(OLD.version);
END IF;
END IF;

RETURN NEW;
END;
$$
LANGUAGE plpgsql;
