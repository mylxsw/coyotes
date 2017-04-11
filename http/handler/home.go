package handler

import (
	"net/http"

	"github.com/mylxsw/coyotes/config"
	"html/template"
	"strconv"
)

const htmlTemplate = `
<html>
<head>
	<title>Coyotes</title>
	<style type="text/css">
	.container {
		width: 500px;
		margin: 0 auto;
	}
	.container .welcome{
		width: 350px;
		margin: 0 auto;
	}
	.container .info {
		width: 480px;
		margin: 0 auto;
	}

	table thead th {
		background-color: rgb(81, 130, 187);
		color: #fff;
		border-bottom-width: 0;
	}
	table td {
		color: #000;
	}
	table tr, table th {
		border-width: 1px;
		border-style: solid;
		border-color: rgb(81, 130, 187);
	}

	table td, table th {
		padding: 5px 10px;
		font-size: 12px;
		font-family: Verdana;
		font-weight: bold;
	}
	</style>
</head>
<body>
	<div class="container">
		<div class="welcome">
			<pre>{{.Welcome}}</pre>
		</div>
		<div class="info">
		  <table>
			<thead>
			  <tr>
				<th style="width: 180px">KEY</th>
				<th style="width: 280px">VALUE</th>
			  </tr>
			</thead>
			<tbody>
			  {{range $k, $v := .Items}}
			  <tr>
				<td>{{$k}}</td>
				<td>{{$v}}</td>
			  </tr>
			  {{end}}
			</tbody>
		  </table>
		</div>
	</div>
</body>
</html>
`

func Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	runtime := config.GetRuntime()

	tmpl, _ := template.New("info").Parse(htmlTemplate)
	tmpl.Execute(w, struct {
		Welcome string
		Items   map[string]string
	}{
		Welcome: config.WelcomeMessage(),
		Items: map[string]string{
			"版本":     config.VERSION,
			"启动时间":   runtime.Info.StartedAt.Format("2006-01-02 15:04:05"),
			"已执行任务数": strconv.Itoa(runtime.Info.DealTaskCount),
			"成功任务数":  strconv.Itoa(runtime.Info.SuccTaskCount),
			"失败任务数":  strconv.Itoa(runtime.Info.FailTaskCount),
			"队列数目":   strconv.Itoa(len(runtime.Channels)),
		},
	})

}
