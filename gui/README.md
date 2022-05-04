# Gui

In package.json I defined this:
```
"homepage": ".",
"proxy": "http://localhost:8082",
```

`homepage` is required to resolve paths relatively as described here: https://create-react-app.dev/docs/deployment/#building-for-relative-paths
`proxy` is used to enable proxy when you execute `npm start`. If you have both the
