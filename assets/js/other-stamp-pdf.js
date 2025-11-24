// $(document).ready(function () {
//     alert('hello')
// });

let file = null;
$('#pdfFileInput').on('change', function() {
    const fileInput = $('#pdfFileInput')[0];
    if (fileInput.files.length > 0) {
        file = fileInput.files[0];
        const fileURL = URL.createObjectURL(file);
        $('#pdfPreview').attr("hidden",false)
        $('#pdfPreview').attr('src', fileURL).show(); // Set the src of iframe and show it
    }
});

function uploadPdfForStamping(){
    if(file == null){
        alert("upload a pdf file!!")
    }else{
        let formData = new FormData();
        formData.append(
            "authenticity_token",
            document.querySelectorAll('[name="authenticity_token"]')[0].value
        );
        formData.append("file", file);
        formData.append("file_name", file.name);
        $.ajax({
            type: "POST",
            url: "/other-pdf-stampify",
            data: formData,
            async: false,
            cache: false,
            contentType: false,
            enctype: "multipart/form-data",
            processData: false,
            success: function (response) {
                console.log(response)
                alert(response.message)
                window.open(response.url, '_blank', 'location=no,menubar=no,titlebar=no, toolbar=no, scrollbars=yes, resizable=yes, top=200, left=500, width=700, height=500')
            },
            error: function (err) {
                //console.log(err)
                $("#uploadBtn").prop("disabled", false);
                $("#uploadBtn").html("Generate Stamp on PDF");
                alert(err.responseText);
            },
        });
    }
}

$('#uploadBtn').click(function (e) {
    e.preventDefault();
    $("#uploadBtn").prop("disabled", true);
    $("#uploadBtn").html("Please Wait...");

    setTimeout(uploadPdfForStamping(),1000)
    $("#uploadBtn").prop("disabled", false);
    $("#uploadBtn").html("Generate Stamp on PDF");
});