ALTER TABLE companies ADD COLUMN in_cart BOOLEAN DEFAULT 0;
ALTER TABLE companies ADD COLUMN cart_added_at DATETIME;
ALTER TABLE companies ADD COLUMN last_notified_at DATETIME;
