SCHEMA=/docker-entrypoint-initdb.d/10-schema.sql
TRIGGER=/trigger.d/db.trigger
[ ! -e $TRIGGER -o $SCHEMA -nt $TRIGGER ] && touch $TRIGGER || true
