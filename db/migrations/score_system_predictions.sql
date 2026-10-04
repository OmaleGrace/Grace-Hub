ALTER TABLE system_predictions ADD COLUMN score NUMERIC;

UPDATE system_predictions sp
SET score = 1 - power(
    sp.confidence - (CASE WHEN lower(trim(sp.predicted)) = lower(trim(e.outcome)) THEN 1 ELSE 0 END), 2)
FROM events e
WHERE e.id = sp.event_id AND e.status = 'resolved' AND e.outcome IS NOT NULL;