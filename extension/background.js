chrome.tabs.onUpdated.addListener(function(tabId, changeInfo, tab) {
    console.log(changeInfo.url);
    if (changeInfo.url == "https://eecsoh.eecs.umich.edu/") {
        chrome.cookies.get({"name": "session", "url": "https://eecsoh.eecs.umich.edu/"}, (cookie) => {
            console.log(cookie)
        })
    }
});

chrome.action.onClicked.addListener((tab) => {
    chrome.tabs.create({
        url: "https://eecsoh.eecs.umich.edu/api/oauth2login"
    });
});