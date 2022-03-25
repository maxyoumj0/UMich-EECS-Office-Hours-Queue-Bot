chrome.tabs.onUpdated.addListener(function(tabId, changeInfo, tab) {
    console.log(changeInfo.url);
    if (changeInfo.url == "https://eecsoh.eecs.umich.edu/") {
        chrome.cookies.get({"name": "session", "url": "https://eecsoh.eecs.umich.edu/"}, (cookie) => {
            console.log(cookie)
            var formdata = new FormData();
            formdata.append("session", cookie.value);

            var requestOptions = {
              method: 'POST',
              body: formdata,
              redirect: 'follow'
            };

            fetch("http://localhost:8081/send_session_eecsoh/", requestOptions)
              .then(response => response.text())
              .then(result => console.log(result))
              .catch(error => console.log('error', error));
                    })
                }

    if (changeInfo.url == "https://oh.eecs.umich.edu/courses") {
        chrome.cookies.getAll({domain: "oh.eecs.umich.edu"}, (cookie1) => {
            console.log(cookie1)
            var formdata = new FormData();
            formdata.append(cookie1[0].name, cookie1[0].value);
            formdata.append(cookie1[1].name, cookie1[1].value);
            console.log(formdata)

            var requestOptions = {
              method: 'POST',
              body: formdata,
              redirect: 'follow'
            };

            fetch("http://localhost:8081/send_session_oh/", requestOptions)
              .then(response => response.text())
              .then(result => console.log(result))
              .catch(error => console.log('error', error));
        })
    }
});