package main

type StaticFileDesc struct {
	ContentType string
	Data        string
}

var staticFiles map[string]StaticFileDesc

const indexHTMLTemplate = `
{{define "html"}}
<!doctype html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Steneteg</title>
    <link rel="stylesheet" href="/static/theme.css">
    <!--<link rel='stylesheet' href='https://fonts.googleapis.com/css?family=Roboto'>-->
    <link href="https://fonts.googleapis.com/css?family=Indie+Flower" rel="stylesheet"> 
    <style>
        html,body,h1,h2,h3,h4,h5,h6 {
            /*font-family: "Roboto", sans-serif;*/
            font-family: 'Indie Flower', cursive;            
        }
    </style>
</head>
<body>
    <div class="title">
        <h1>{{ .title }}</h1>
        <p>{{ .headline }}</p>
    </div>
    <script src="/static/app.js"></script>
</body>
</html>
{{end}}
`

const themeCSSStatic = `
html, body {
    margin: 0;
    padding: 0;
    height: 100%;
    position: relative;
    background-color: #535353;
}
div.title {
    margin: 0;
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
}
h1 {
    margin: 0;
    text-align: center;
    font-size: 25vmin;
    text-shadow: 4px 6px 8px #000;
    line-height: 1;
}

p {
    margin: 0;
    text-align: center;
    font-size: 6vmin;
    /*text-shadow: 4px 6px 8px #000;*/
}
`

func init() {
	staticFiles = map[string]StaticFileDesc{}
	staticFiles["theme.css"] = StaticFileDesc{"text/css", themeCSSStatic}
}
