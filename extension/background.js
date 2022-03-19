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

            fetch("http://localhost:3000/send_session/", requestOptions)
              .then(response => response.text())
              .then(result => console.log(result))
              .catch(error => console.log('error', error));
                    })
                }
});

chrome.action.onClicked.addListener((tab) => {
    chrome.tabs.create({
        url: "https://eecsoh.eecs.umich.edu/api/oauth2login"
    });
});