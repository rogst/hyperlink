var app = new Vue({
	delimiters: ['${','}'],
	el: '#app',
	data: {
		error: '',
		hyperlink: '',
		secretMessage: '',
		viewsLeft: '',
		expireIn: ''
	},
	methods: {
		getHyperlink: function(event) {
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
					console.log(this.responseText);
					that.error = this.responseText;
				}
			}

			xhr.send();
		}
	}
})

function converBase64toBlob(content) {
	var sliceSize = 512;
	var byteCharacters = window.atob(content); //method which converts base64 to binary
	var byteArrays = [
	];
	for (var offset = 0; offset < byteCharacters.length; offset += sliceSize) {
	  var slice = byteCharacters.slice(offset, offset + sliceSize);
	  var byteNumbers = new Array(slice.length);
	  for (var i = 0; i < slice.length; i++) {
		byteNumbers[i] = slice.charCodeAt(i);
	  }
	  var byteArray = new Uint8Array(byteNumbers);
	  byteArrays.push(byteArray);
	}
	var blob = new Blob(byteArrays, {}); //statement which creates the blob
	return blob;
  }