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
						that.expireIn = secsToDuration(that.hyperlink.expireIn / 1000000000);
					}
				} else {
					that.error = this.responseText;
				}
			}

			xhr.send();
		}
	}
})

function secsToDuration(seconds) {
	dur = ""
	if (seconds >= (60 * 60 * 24)) {
		days = Math.trunc(seconds / (60 * 60 * 24))
		dur += days + "d ";
		seconds -= (days * 60 * 60 * 24)
	}
	if (seconds >= (60 * 60)) {
		hours = Math.trunc(seconds / (60 * 60))
		dur += hours + "h "
		seconds -= (hours * 60 * 60)
	}
	if (seconds >= 60) {
		mins = Math.trunc(seconds / 60)
		dur += mins + "m "
		seconds -= (mins * 60)
	}
	secs = seconds % 60
	if (secs > 0) {
		dur += secs + "s"
	}
	return dur.trim()
}