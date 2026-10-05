ALTER TABLE events ADD COLUMN options TEXT[] NOT NULL DEFAULT ARRAY['home','draw','away'];

UPDATE events SET options = ARRAY['above','below'] WHERE category = 'stocks';

ALTER TABLE events ADD CONSTRAINT events_options_min CHECK (array_length(options, 1) >= 2);