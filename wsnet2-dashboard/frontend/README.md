# Frontend

[日本語](README-ja.md)

SPA built using `Vue 3 + Typescript + Vite`. [NaiveUI](https://www.naiveui.com/en-US/os-theme) is used for UI.

## Environment variables

| Name                    | Description                  | Example                 |
| ----------------------- | ---------------------------- | ----------------------- |
| VITE_DEFAULT_SERVER_URI | Default dashboard server uri | "http://localhost:5555" |

## Development environment

- [VSCode](https://code.visualstudio.com/) + [Volar](https://marketplace.visualstudio.com/items?itemName=johnsoncodehk.volar)

## Commands

- `npm run dev`：run development server
- `npm run build`：build the application（stored at `./dist`）

## Build the application with Docker

```bash
cd wsnet2-dashboard
docker compose run --rm frontbuilder
```

The built code will be stored at `wsnet2-dashboard/frontend/dist`.
