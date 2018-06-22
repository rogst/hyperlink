var app = new Vue({
	delimiters: ['${','}'],
	el: '#app',
	data: {
		hyperlink: '',
		err: '',
		showClearFileBtn: false
	},
	methods: {
		getHyperlink: function(event) {
			event.preventDefault();
			var xhr = new XMLHttpRequest();
			xhr.open('POST', '/api/', true);
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

			if (document.getElementById("secretFile").files.length == 1) {
				var formData = new FormData();
				var file = document.getElementById("secretFile").files[0];
				formData.append('data', file, file.name);
				formData.append('maxViews', document.getElementById("maxViews").value);
				formData.append('expireIn', document.getElementById("expireIn").value);
				xhr.send(formData);
			} else {
				xhr.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
				var params = "data=" + document.getElementById("secretMessage").value;
				params += "&maxViews=" + document.getElementById("maxViews").value;
				params += "&expireIn=" + document.getElementById("expireIn").value;
				xhr.send(params);
			}
		},
		messageUpdate: function (event) {
			if (document.getElementById("secretMessage").value.length > 0) {
				document.getElementById("secretFile").disabled = true;
			} else {
				document.getElementById("secretFile").disabled = false;
			}
		},
		fileUpdate: function (event) {
			if (document.getElementById("secretFile").files.length > 0) {
				document.getElementById("secretMessage").disabled = true;
				this.showClearFileBtn = true;
			} else {
				document.getElementById("secretMessage").disabled = false;
				this.showClearFileBtn = false;
			}
		},
		clearFile: function(event) {
			document.getElementById("secretFile").value = null;
		}
	}
})