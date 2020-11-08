let submitBtn = <HTMLInputElement> document.querySelector("#submit-btn");
let getPostsBtn = <HTMLInputElement> document.querySelector("#get-posts-btn");

function getRecentPosts(page = 1) {
    let req = new XMLHttpRequest();

    req.addEventListener("load", function(event) {
        let data = JSON.parse(req.responseText);
        console.log(data);
        
        // Get only the top 5 posts
        data.results = data.results.slice(0,5);
        updateViewMain(data);
    });

    req.open("GET", "/api/v1/posts?page=" + page);
    req.send();
}

function updateViewMain(data: any[]) {
    // Get the template from the DOM.
    let template = document.querySelector("#recent-posts-template").innerHTML;

    // Create a render function for the template with doT.template.
    let renderFunction = doT.template(template);

    // Use the render function to render the data.
    let rendered = renderFunction(data);

    // Insert the result into the DOM (inside the <div> with the ID log.
    document.querySelector("#recent-posts").innerHTML = rendered;
}

function createPost() {
    let titleInput : HTMLInputElement = document.querySelector("#title-input");
    let textInput : HTMLInputElement = document.querySelector("#content-input");
    let publicInput : HTMLInputElement = document.querySelector("#public-input");

    let req = new XMLHttpRequest();

    req.addEventListener("load", function() {
        let data = JSON.parse(req.responseText);
        window.location.href = "/posts/" + data.postContent.writeId;
    });

    req.open("POST", "/api/v1/posts");
    req.send(JSON.stringify({
        title: titleInput.value,
        text: textInput.value,
        public: publicInput.checked,
    }));
}

function attachMainButtonlisteners() {
    submitBtn.addEventListener("click", createPost);
    getPostsBtn.addEventListener("click", function() {
        window.location.href = '/posts';
    });
}

getRecentPosts();
attachMainButtonlisteners();