function showReportForm() {
    // Get the template from the DOM.
    let template = document.querySelector("#report-form-template").innerHTML;

    // Create a render function for the template with doT.template.
    let renderFunction = doT.template(template);

    // Use the render function to render the data.
    let rendered = renderFunction();

    // Insert the result into the DOM (inside the <div> with the ID log.
    document.querySelector("#report-form").innerHTML = rendered;

    // Add event listener to the newly-added button
    let reportSubmitBtn = <HTMLInputElement> document.querySelector("#report-submit-btn");
    reportSubmitBtn.addEventListener("click", submitReportForm);
}

function submitReportForm() {
    let reason: HTMLInputElement = document.querySelector('input[name="report-reason"]:checked');
    createReport(reason.value);
}

function createReport(reason: string) {
    let req = new XMLHttpRequest();

    req.addEventListener("load", function(event) {
        alert("Post reported!");
        window.location.href = '/';
    });

    req.open("POST", "/api/v1" + window.location.pathname + "/reports");
    req.send(JSON.stringify({
        reason: reason,
    }));
}

let reportBtn = <HTMLInputElement> document.querySelector("#report-btn");
reportBtn.addEventListener("click", showReportForm);
