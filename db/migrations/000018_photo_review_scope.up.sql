-- Photo review applies to the reviewed public photo set, not future replacements.
CREATE FUNCTION invalidate_matrimony_photo_review() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE who TEXT; changed BOOLEAN;
BEGIN
 IF TG_OP='INSERT' THEN
  changed:=NEW.published; who:=NEW.owner;
 ELSIF TG_OP='DELETE' THEN
  changed:=OLD.published; who:=OLD.owner;
 ELSE
  changed:=OLD.published IS DISTINCT FROM NEW.published
    OR ((OLD.published OR NEW.published) AND (OLD.data IS DISTINCT FROM NEW.data OR OLD.owner IS DISTINCT FROM NEW.owner));
  who:=OLD.owner;
 END IF;
 IF changed THEN
  UPDATE matrimony_profiles SET verified_at=NULL WHERE account_id=who;
  UPDATE matrimony_verifications SET status='rejected',note='Published photos changed. Submit a new selfie for review.',photo=NULL
   WHERE account_id=who AND status='approved';
  IF TG_OP='UPDATE' AND OLD.owner IS DISTINCT FROM NEW.owner THEN
   UPDATE matrimony_profiles SET verified_at=NULL WHERE account_id=NEW.owner;
   UPDATE matrimony_verifications SET status='rejected',note='Published photos changed. Submit a new selfie for review.',photo=NULL
    WHERE account_id=NEW.owner AND status='approved';
  END IF;
 END IF;
 RETURN NULL;
END $$;
CREATE TRIGGER matrimony_photo_review_scope AFTER INSERT OR UPDATE OR DELETE ON matrimony_photos
 FOR EACH ROW EXECUTE FUNCTION invalidate_matrimony_photo_review();
