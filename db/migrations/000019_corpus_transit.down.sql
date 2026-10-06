DELETE FROM astro_corpus WHERE doc_type='transit';
ALTER TABLE astro_corpus DROP CONSTRAINT astro_corpus_doc_type_check,
 ADD CONSTRAINT astro_corpus_doc_type_check CHECK (doc_type IN ('yoga','graha_in_house','graha_in_sign','nakshatra','dasha','dignity','bhava','aspect','karaka'));
