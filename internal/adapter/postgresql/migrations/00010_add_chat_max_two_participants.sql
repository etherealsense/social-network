-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION check_chat_participants_limit()
RETURNS TRIGGER AS $$
BEGIN
    IF (SELECT COUNT(*) FROM chat_participants WHERE chat_id = NEW.chat_id) >= 2 THEN
        RAISE EXCEPTION 'chat cannot have more than 2 participants';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_chat_participants_limit
BEFORE INSERT ON chat_participants
FOR EACH ROW EXECUTE FUNCTION check_chat_participants_limit();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_chat_participants_limit ON chat_participants;
DROP FUNCTION IF EXISTS check_chat_participants_limit();
-- +goose StatementEnd
