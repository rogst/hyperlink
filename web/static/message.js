Vue.createApp({
	delimiters: ['${', '}'],
	data() {
		return {
			error: '',
			message: '',
		}
	},
	methods: {
		openMessage: function (event, msgType, link) {
			event.preventDefault();
			if (msgType === "file") {
				window.location = link;
			} else {
				that = this;
				var xhr = new XMLHttpRequest();
				xhr.open('GET', link, true);
				xhr.onload = function () {
					if (this.status == 200) {
						that.message = this.responseText;
					} else {
						that.error = this.responseText;
					}
				}
				xhr.send();
			}
		}
	}
}).mount('#app')
