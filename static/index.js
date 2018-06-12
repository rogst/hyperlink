var app = new Vue({
	delimiters: ['${','}'],
	el: '#app',
	data: {
		hyperlink: '',
		err: ''
	},
	methods: {
		getHyperlink: function(event) {
			event.preventDefault();
			var xhr = new XMLHttpRequest();
			xhr.open('POST', '/api/', true);
			xhr.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
			var params = "secretMessage=" + document.getElementById("secretMessage").value;
			params += "&maxViews=" + document.getElementById("maxViews").value;
			params += "&expireIn=" + document.getElementById("expireIn").value;
			that = this;

			xhr.onload = function() {
				if (this.status == 200) {
					that.hyperlink = window.location.href + this.responseText;
					that.err = '';
				} else {
					that.err = this.responseText;
				}
			}
  			xhr.onerror = function(){ console.log("hyperlink request failed"); }

			xhr.send(params);
		}
	}
})