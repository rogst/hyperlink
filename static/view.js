var app = new Vue({
	delimiters: ['${','}'],
	el: '#app',
	data: {
	  	hyperlink: ''
	},
	methods: {
		getHyperlink: function(event) {
			event.preventDefault();
			var xhr = new XMLHttpRequest();
			xhr.open('GET', '/api/'+window.location.pathname, true);
			that = this;

			xhr.onload = function() {
				that.hyperlink = this.responseText;
			}
  			xhr.onerror = function(){ console.log("hyperlink request failed"); }

			xhr.send();
		}
	}
})