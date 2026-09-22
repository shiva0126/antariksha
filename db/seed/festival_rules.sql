INSERT INTO festival_rules(slug,display_name,rule_type,tithi,paksha,region,notes) VALUES
 ('shukla-ekadashi','Shukla Ekadashi','tithi',11,'shukla','all','Observed monthly'),
 ('krishna-ekadashi','Krishna Ekadashi','tithi',11,'krishna','all','Observed monthly'),
 ('purnima','Purnima','tithi',15,'shukla','all','Full moon observance'),
 ('amavasya','Amavasya','tithi',15,'krishna','all','New moon observance')
ON CONFLICT(slug) DO NOTHING;

