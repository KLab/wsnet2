#!/bin/sh
npm run generate &&
npm run build &&
node dist/index.js