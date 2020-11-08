let metadata = {
    page: 1,
    perPage: 20,
    totalShowing: 0,
    totalCount: 0,
    totalPages: 1,
}

let prevBtn = <HTMLInputElement> document.querySelector("#previous-page-btn");
let nextBtn = <HTMLInputElement> document.querySelector("#next-page-btn");

function getPosts(page: number) {
    let req = new XMLHttpRequest();

    req.addEventListener("load", function(event) {
        let data = JSON.parse(req.responseText);
        console.log(data);
        metadata = data.metadata;
        console.log(metadata);
        updateButtons();
        updateView(data);
    });

    req.open("GET", "/api/v1/posts?page=" + page);
    req.send();
}

function updateView(data: any[]) {
    // Get the template from the DOM.
    let template = document.querySelector("#posts-template").innerHTML;

    // Create a render function for the template with doT.template.
    let renderFunction = doT.template(template);

    // Use the render function to render the data.
    let rendered = renderFunction(data);

    // Insert the result into the DOM (inside the <div> with the ID log.
    document.querySelector("#posts").innerHTML = rendered;
}



function updateButtons() {
    prevBtn.disabled = (metadata.page == 1);
    nextBtn.disabled = (metadata.page == metadata.totalPages);
}

function getPreviousPage() {
    metadata.page -= 1;
    getPosts(metadata.page);
}

function getNextPage() {
    metadata.page += 1;
    console.log("hi");
    getPosts(metadata.page);
}

function attachButtonListeners() {
    prevBtn.addEventListener("click", getPreviousPage);
    nextBtn.addEventListener("click", getNextPage);
}

getPosts(metadata.page);
attachButtonListeners();