CREATE OR REPLACE FUNCTION create_portfolio_on_true()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.portfolio THEN
        IF NOT EXISTS (SELECT 1 FROM portfolio WHERE finder_id = NEW.id) THEN
            INSERT INTO portfolio (finder_id) VALUES (NEW.id);
        END IF;
    ELSE
        DELETE FROM portfolio WHERE finder_id = NEW.id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_create_portfolio_on_true
AFTER INSERT OR UPDATE ON finders
FOR EACH ROW
EXECUTE FUNCTION create_portfolio_on_true();