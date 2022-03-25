eecsoh.addEventListener("click", async()=>{
    chrome.tabs.create({
        url: "https://eecsoh.eecs.umich.edu/api/oauth2login"
    });
});

oh.addEventListener("click", async()=>{
    chrome.tabs.create({
        url: "https://oh.eecs.umich.edu/auth/google_oauth2?origin=https://oh.eecs.umich.edu/courses"
    });
});