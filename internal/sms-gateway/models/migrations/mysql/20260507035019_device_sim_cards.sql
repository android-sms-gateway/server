-- +goose Up
-- +goose StatementBegin
ALTER TABLE `devices`
ADD `sim_cards` json NOT NULL DEFAULT ('[]');
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
ALTER TABLE `devices` DROP `sim_cards`;
-- +goose StatementEnd