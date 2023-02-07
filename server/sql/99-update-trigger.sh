#!/bin/bash
SCHEMA=/docker-entrypoint-initdb.d/10-schema.sql
TRIGGER=/trigger.d/db.trigger
[[ ! -e $TRIGGER || $SCHEMA -nt $TRIGGER ]] && touch $TRIGGER || true
