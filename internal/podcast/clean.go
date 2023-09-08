package podcast

import "strings"

var acast string = `<hr><p style='color:grey; font-size:0.75em;'> Hosted on Acast. See <a style='color:grey;' target='_blank' rel='noopener noreferrer' href='https://acast.com/privacy'>acast.com/privacy</a> for more information.</p>`
var art19 string = `<p>Unsere allgemeinen Datenschutzrichtlinien finden Sie unter <a href="https://art19.com/privacy" rel="noopener noreferrer" target="_blank">https://art19.com/privacy</a>. Die Datenschutzrichtlinien für Kalifornien sind unter <a href="https://art19.com/privacy#do-not-sell-my-info" rel="noopener noreferrer" target="_blank">https://art19.com/privacy#do-not-sell-my-info</a> abrufbar.</p>`
var art19_2 string = `Unsere allgemeinen Datenschutzrichtlinien finden Sie unter https://art19.com/privacy. Die Datenschutzrichtlinien für Kalifornien sind unter https://art19.com/privacy#do-not-sell-my-info abrufbar.`
var p_br_p string = `<p><br></p>`

func CleanDescription(s string) string {
	s = strings.ReplaceAll(s, acast, "")
	s = strings.ReplaceAll(s, art19, "")
	s = strings.ReplaceAll(s, art19_2, "")
	s = strings.ReplaceAll(s, p_br_p, "")
	return s
}
