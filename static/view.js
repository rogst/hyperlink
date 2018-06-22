var app = new Vue({
	delimiters: ['${','}'],
	el: '#app',
	data: {
		error: '',
		hyperlink: '',
		viewsLeft: '',
		expireIn: ''
	},
	methods: {
		showHyperlink: function(event) {
			event.preventDefault();
			var xhr = new XMLHttpRequest();
			xhr.open('GET', '/api'+window.location.pathname, true);
			that = this;

			xhr.onload = function() {
				if (this.status == 200) {
					that.hyperlink = JSON.parse(this.responseText);
					that.viewsLeft = that.hyperlink.maxViews - that.hyperlink.views;
					if (that.viewsLeft > 0) {
						var seconds = that.hyperlink.expireIn / 1000000000;
						that.expireIn = seconds + 's';
					}
				} else {
					that.error = this.responseText;
				}
			}

			xhr.send();
		}
	}
})
