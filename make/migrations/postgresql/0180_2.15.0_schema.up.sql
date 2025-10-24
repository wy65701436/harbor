/*
Initialize skip_audit_log_database configuration based on existing audit log usage.

Logic:
1. Skip if configuration already exists
2. Set to 'false' if:
   - Any audit logs exist in audit_log or audit_log_ext tables, OR
   - Tables show evidence of previous usage (sequence value > 1)
3. Otherwise, don't create the configuration (defaults to true in code)
*/
DO $$
DECLARE
    has_audit_logs BOOLEAN;
    has_table_usage BOOLEAN;
BEGIN
    -- Exit early if configuration already exists
    IF EXISTS (SELECT 1 FROM properties WHERE k = 'skip_audit_log_database') THEN
        RETURN;
    END IF;

    -- Check if any audit logs exist
    has_audit_logs := EXISTS (SELECT 1 FROM audit_log LIMIT 1) 
                   OR EXISTS (SELECT 1 FROM audit_log_ext LIMIT 1);

    -- Check if tables have been used (sequence value > 1 indicates usage)
    has_table_usage := (SELECT last_value FROM audit_log_id_seq) > 1
                    OR (SELECT last_value FROM audit_log_ext_id_seq) > 1;

    -- Insert configuration only if audit logs exist or tables have been used
    IF has_audit_logs OR has_table_usage THEN
        INSERT INTO properties (k, v) VALUES ('skip_audit_log_database', 'false');
    END IF;
END $$;
