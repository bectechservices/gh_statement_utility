
let finalEndDate = null
let srvStartDate = ""
let srvEndDate = ""

function CheckWhetherInputIsInSrvDateRange(startDate,endDate){
    const start = Date.parse(startDate)
    const end  = Date.parse(endDate)

    const srvStart =  Date.parse(srvStartDate)
    const srvEnd =  Date.parse(srvEndDate)

    if (start < srvStart || start > srvEnd){
        return false
    }
    else if (end > srvEnd){
        return false
    }
    return true

//     console.log("start: ",start)
//     console.log("end: ",end)
//     console.log("srvStartDate: ",srvStart)
//     console.log("srvEndDate: ",srvEnd)
//    return start <=srvEnd && end <= srvEnd
}

//when validate button is clicked
$('#validateBtn').on("click", function(e) {
    clearErrorTexts()
    clearCustomerDetails();
    //validate input fields
    let accountNum = $('#accountNumber').val();
    let startDate = $('#startDate').val();
    let endDate = $('#endDate').val();

    // if (!CheckWhetherInputIsInSrvDateRange(startDate,endDate)){
    //     alert(`Date input must fall between ${srvStartDate} and ${srvEndDate}`)
    //     return
    // }

    // if all fields are validated and successful
    let valAccNum = validateAccountNum(accountNum)
    let valInputDates = validateInputDates(startDate, endDate)
    if (valAccNum && valInputDates) {
        $('#tableCard').html(loadingContent);
        let token = document.querySelector('meta[name="csrf-token"]').getAttribute('content');
        let url = '/validate';
        let data = {
            authenticity_token: token,
            account_no: accountNum,
            start_date: startDate,
            end_date: endDate,
        };

        $.ajax({
            type: "POST",
            url: url,
            data: data,
            dataType: "JSON",
            success: function(response) {
                $('#tableCard').html(displayValTable(response))

            },
            error: function(error) {
                //alert("BECTECH. Something went wrong");
                //$('#tableCard').html("")

                if(error.status == 200)
                {
                    let errRes = error.responseText;
                    let data = errRes.split('<!DOCTYPE html>')[0]
                    console.log(data)
                    $('#tableCard').html(displayValTable(JSON.parse(data)))

                }
                else{
                    alert("BECTECH. Something went wrong");
                    $('#tableCard').html("")
                }
            }
        });

    }
});

//search button
$('#searchBtn').on('click', function(e) {
    //validate input fields
    let accountNum = $('#accountNumber').val();
    let startDate = $('#startDate').val();
    let endDate = $('#endDate').val();

    if (!CheckWhetherInputIsInSrvDateRange(startDate,endDate)){
        alert(`Date input must fall between ${formatDate(srvStartDate)} and ${formatDate(srvEndDate)}`)
        return
    }

    // if all fields are validated and successful
    if (validateAccountNum(accountNum) && validateInputDates(startDate, endDate)) {
        $('#tableCard').html(loadingContent);
        let token = document.querySelector('meta[name="csrf-token"]').getAttribute('content');
        let url = '/search';
        let data = {
            authenticity_token: token,
            account_no: accountNum,
            start_date: startDate,
            end_date: endDate,
        };

        $.ajax({
            type: "POST",
            url: url,
            data: data,
            dataType: "JSON",
            success: function(response) {
                $('#tableCard').html(displaySearchTable(response))

            },
            error: function(error) {
                //alert("BECTECH. Something went wrong");
                //$('#tableCard').html("")

                if(error.status == 200)
                {
                    let errRes = error.responseText;
                    let data = errRes.split('<!DOCTYPE html>')[0]
                    console.log(data)
                    $('#tableCard').html(displaySearchTable(JSON.parse(data)))

                }
                else{
                    alert("BECTECH. Something went wrong");
                    $('#tableCard').html("")
                }
            }
        });

    }
});

$('#accountNumber').on("blur", function(e) {
    if ($('#accountNumber').val() != '') {
        $('#startDate').val('');
        $('#endDate').val('');
        $('#tableCard').html('');

        let accountNum = $('#accountNumber').val();
        let token = document.querySelector('meta[name="csrf-token"]').getAttribute('content');
        $.ajax({
            type: "POST",
            url: "/account-start",
            data: {
                authenticity_token: token,
                account_no: accountNum
            },
            dataType: "json",
            success: function(response) {
                if (response != 0) {
                    if (response) {
                        let startDay = response.Day;
                        let startMonth = response.Month;
                        let startYear = response.Year;
                        let endDay = response.ltable_d;
                        let endMonth = response.ltable_m;
                        let endYear = response.ltable_y;

                        clearCustomerDetails();
                        if (startYear == '' || startYear == null) {
                            //alert('Account number not found');
                            //$('#displayPdfBtn').attr('disabled', true);
                            $('#searchBtn').attr('disabled', true);
                            $('#validateBtn').attr('disabled', true);
                            $('#printBtn').attr('disabled', true);
                            $('#excelBtn').attr('disabled', true);
                            $('#startDate').attr('disabled', true);
                            $('#endDate').attr('disabled', true);

                            $('#tableCard').html(displayAccountDoesNotExist);
                        } else {
                            $('#searchBtn').attr('disabled', false);
                            $('#validateBtn').attr('disabled', false);
                            $('#printBtn').attr('disabled', false);
                            $('#excelBtn').attr('disabled', false);
                            $('#startDate').attr('disabled', false);
                            $('#endDate').attr('disabled', false);


                            document.getElementById("startDate").value = new Date(startYear, startMonth - 1, startDay).toISOString().split('T')[0];
                            document.getElementById("startDate").setAttribute('min', new Date(startYear, startMonth - 1, startDay).toISOString().split('T')[0]);
                            document.getElementById("startDate").setAttribute('max', new Date(endYear, endMonth - 1, endDay).toISOString().split('T')[0]);

                            document.getElementById("endDate").value = new Date(endYear, endMonth - 1, endDay).toISOString().split('T')[0];
                            document.getElementById("endDate").setAttribute('min', new Date(startYear, startMonth - 1, startDay).toISOString().split('T')[0]);
                            document.getElementById("endDate").setAttribute('max', new Date(endYear, endMonth - 1, endDay).toISOString().split('T')[0]);
                            //$('#displayPdfBtn').attr('disabled', false);

                            srvStartDate = document.getElementById("startDate").value
                            srvEndDate = document.getElementById("endDate").value

                        }


                    }


                }
            }
        });
    }
});

//displa pdf button clicked
$('#printBtn').on('click', function(e) {
    //validate input fields
    let accountNum = $('#accountNumber').val();
    let startDate = $('#startDate').val();
    let endDate = $('#endDate').val();
    let currency = $('#currency').html();

    // if all fields are validated and successful
    if (validateAccountNum(accountNum) && validateInputDates(startDate, endDate)) {
        //$('#tableCard').html(loadingContent);
        //store session
        window.sessionStorage.setItem('account_no',accountNum)
        window.sessionStorage.setItem('start_date',startDate)
        window.sessionStorage.setItem('end_date',endDate)

        //alert(sessionStorage.getItem('account_no'))

        let token = document.querySelector('meta[name="csrf-token"]').getAttribute('content');
        let url = `/print?account_no=${accountNum}&start_date=${startDate}&end_date=${endDate}`;

        window.open(url, '_blank', 'location=no,menubar=no,titlebar=no, toolbar=no, scrollbars=yes, resizable=yes, top=200, left=500, width=700, height=500')
    }
});

$('#excelBtn').on('click', function(e) {
    //validate input fields
    let accountNum = $('#accountNumber').val();
    let startDate = $('#startDate').val();
    let endDate = $('#endDate').val();
    let currency = $('#currency').html();

    // if all fields are validated and successful
    if (validateAccountNum(accountNum) && validateInputDates(startDate, endDate)) {
        let token = document.querySelector('meta[name="csrf-token"]').getAttribute('content');
        let url = `/excel?account_no=${accountNum}&start_date=${startDate}&end_date=${endDate}`;
        window.open(url, '_blank', 'location=no,menubar=no,titlebar=no, toolbar=no, scrollbars=yes, resizable=yes, top=200, left=500, width=700, height=500')
    }
});

function validateAccountNum(accountNum) {
    if (accountNum == '') {
        $('#accno_err').html('Account no. required');
        return false;
    } else if (Number(accountNum) == false) {
        $('#accno_err').html('Provide a valid account number.');
        return false;
    } else {
        return true;
    }
}

function validateInputDates(startDate, endDate) {
    if (startDate == '') {
        $('#start_err').html('Start date required');
        return false;
    } else if (endDate == '') {
        $('#end_err').html('End date required');
        return false;
    } else {
        return true;
    }
}

function SentModalContents(oldStartDate,newStartDate) {
    let msg = `
    No transactions were performed or found on your input start date <strong>${formatDate(oldStartDate)}</strong>.
    Please note that the actual transaction activities started on <strong>${formatDate(newStartDate)}</strong>. 
    As a result, the start date and end date fields will be updated to reflect this information.
    `

    $("#modalContent").html(msg);
}

function formatDate(dateString) {
    // Parse the date string (format: YYYY-MM-DD)
    let [year, month, day] = dateString.split('-').map(Number);

    // Convert month number to month name
    const monthNames = [
        "January", "February", "March", "April", "May", "June",
        "July", "August", "September", "October", "November", "December"
    ];

    // Function to get the ordinal suffix for the day
    function getOrdinalSuffix(day) {
        if (day > 3 && day < 21) return 'th'; // Special case for 11th to 20th
        switch (day % 10) {
            case 1: return 'st';
            case 2: return 'nd';
            case 3: return 'rd';
            default: return 'th';
        }
    }

    // Construct the formatted date
    const dayWithSuffix = day + getOrdinalSuffix(day);
    const monthName = monthNames[month - 1];

    return `${dayWithSuffix} ${monthName}, ${year}`;
}


// prettier-ignore
/* prettier-ignore */
const displayValTable = (response) => {
    // if user input start date and db start date different
    const firstNonZeroTransaction = response.find(item => item.total_tnx !== "0.00" || item.opening_balance !== "0.00");
    if (firstNonZeroTransaction == undefined | null){
        $('#tableCard').html(displayMessageInTable("Account/Transactions Does Not exist for this Date Range"));
        return
    }
    //console.log("firstNonZeroTransaction: ",firstNonZeroTransaction)
    let startDate = new Date($('#startDate').val())
    let responseStartDate = new Date(firstNonZeroTransaction.finalStartDate)

    // console.log("startDate: ",startDate)
    // console.log("startDate.getFullYear(): ",startDate.getFullYear())
    // console.log("startDate.getMonth(): ",startDate.getMonth())
    // console.log("responseStartDate: ",responseStartDate)
    // console.log("responseStartDate.getFullYear(): ",responseStartDate.getFullYear())
    // console.log("responseStartDate.getMonth(): ",responseStartDate.getMonth())

    //  if ($('#startDate').val() != response[0].finalStartDate){
    //     SentModalContents($('#startDate').val(), response[0].finalStartDate)
    //     $("#infoModalBtn").click();
    // }

    if (startDate.getFullYear() !== responseStartDate.getFullYear() || startDate.getMonth() !== responseStartDate.getMonth()){
        SentModalContents($('#startDate').val(), firstNonZeroTransaction.finalStartDate)
        $("#infoModalBtn").click();
    }
    $('#startDate').val(firstNonZeroTransaction.finalStartDate)
    $('#endDate').val(response[response.length-1].finalEndDate)
    srvStartDate = response[0].finalStartDate
    srvEndDate = response[response.length-1].finalEndDate

    //table responsive
    let tableResponsiveDiv = document.createElement('div');
    tableResponsiveDiv.classList.add('table-responsive');

    //table
    let validateTable = document.createElement('table');
    validateTable.classList.add('table')
    validateTable.classList.add('table-hover')
    validateTable.classList.add('table-bordered');

    //table head
    let tableHead = document.createElement('thead')
    tableHead.classList.add('thead-dark');
    let tHeadRow = document.createElement('tr');

    //table body
    let tableBody = document.createElement('tbody');

    let headCols = ['Month', 'Opening Balance', 'Total Transactions', 'Closing Balance', 'Valid'];
    //add th to tr
    for (let i = 0; i < headCols.length; i++) {

        let th = document.createElement('th');
        th.setAttribute('scope', 'col');
        th.innerHTML = headCols[i];

        tHeadRow.appendChild(th);
    }

    for (let i = 0; i < response.length; i++) {
        //add elements to tbody
        let tBodyRow = document.createElement('tr');
        let color = '';
        if (response[i].validity == 'FALSE') {
            color = 'text-danger fw-bold';
        }

        tBodyRow.innerHTML += `<td>${response[i].current_month}</td>`;
        tBodyRow.innerHTML += `<td>${response[i].opening_balance}</td>`;
        tBodyRow.innerHTML += `<td>${response[i].total_tnx}</td>`;
        tBodyRow.innerHTML += `<td>${response[i].closing_balance}</td>`;
        tBodyRow.innerHTML += `<td class="${color}">${response[i].validity}</td>`;

        tableBody.append(tBodyRow);
    }

    //tableHead.append(tHeadRow);
    $(tableHead).html(tHeadRow);
    $(validateTable).html(tableHead);
    $(validateTable).append(tableBody);
    $(tableResponsiveDiv).html(validateTable);
    return tableResponsiveDiv;

}

const loadingContent = () => {
    let div = document.createElement('div');
    $(div).addClass('spinner-border text-primary');
    $(div).attr('role', 'status');

    let span = document.createElement('span');
    //$(span).addClass('visually-hidden');
    //$(span).html('Loading.......');

    $(div).html(span);
    let DIV = document.createElement('div');
    $(DIV).html(div);

    return `<div class="d-flex align-items-center justify-content-center p-5">${$(DIV).html()}</div>`;

}

//search table
const displaySearchTable = (response) => {
    //table responsive
    let tableResponsiveDiv = document.createElement('div');
    tableResponsiveDiv.classList.add('table-responsive');

    //table
    let validateTable = document.createElement('table');
    validateTable.classList.add('table')
    validateTable.classList.add('table-hover')
    validateTable.classList.add('table-bordered');

    //table head
    let tableHead = document.createElement('thead')
    tableHead.classList.add('thead-dark');
    let tHeadRow = document.createElement('tr');

    //table body
    let tableBody = document.createElement('tbody');

    let headCols = ['Entry Value', 'Value Date', 'Particulars', 'Withdrawal', 'Deposit', 'Balance'];
    //add th to tr
    for (let i = 0; i < headCols.length; i++) {

        let th = document.createElement('th');
        $(th).addClass("text-center")
        th.setAttribute('scope', 'col');
        th.innerHTML = headCols[i];

        tHeadRow.appendChild(th);
    }

    //check if data has transactions
    let hasTransaction = response.has_transaction;

    if (hasTransaction === true) {
        //get forward balance
        let balBroughtForward = response.bal_brought_forward;
        let transactions = response.data.transaction;


        if(transactions != null || transactions != undefined){
            $(tableBody).html(displayForwardBalance(balBroughtForward));
            for (let i = 0; i < transactions.length; i++) {
                //add elements to tbody
                let tBodyRow = document.createElement('tr');

                tBodyRow.innerHTML += `<td>${transactions[i].entry_date}</td>`;
                tBodyRow.innerHTML += `<td>${transactions[i].value_date}</td>`;
                tBodyRow.innerHTML += `<td>${transactions[i].narration}</td>`;
                tBodyRow.innerHTML += `<td class="text-end" style='text-align:right;'>${transactions[i].debit_amt}</td>`;
                tBodyRow.innerHTML += `<td class="text-end" style='text-align:right;'>${transactions[i].credit_amt}</td>`;
                tBodyRow.innerHTML += `<td class="text-end" style='text-align:right;'>${transactions[i].opening_balance}</td>`;

                tableBody.append(tBodyRow);
            }
        }
    } else {
        let tBodyRow = document.createElement('tr');
        $(tBodyRow).html(`<td colspan="6" class="text-danger fs-5">${response.data.transaction}</td>`);
        tableBody.append(tBodyRow);
    }

    $(tableHead).html(tHeadRow);
    $(validateTable).html(tableHead);
    $(validateTable).append(tableBody);
    $(tableResponsiveDiv).html(validateTable);

    setCustomerDetails([response.data.customer_name, response.data.branch, response.data.currency, response.data.total_credits, response.data.total_debits]);
    return tableResponsiveDiv;
}

const displayForwardBalance = (data) => {
    let tr = document.createElement('tr');
    $(tr).addClass("table-primary");
    tr.innerHTML += `<td colspan="1" class="fw-bold">${data.date}</td>`;
    tr.innerHTML += `<td colspan="4" class="text-center fw-bold">${data.title}</td>`;
    tr.innerHTML += `<td colspan="1" class="fw-bold" style='text-align:right;'>${data.open_balance}</td>`;

    return tr;
}

function setCustomerDetails(data) {

    const [customer_name, branch, currency, total_credits, total_debits] = data;
    $('#customer_name').text(customer_name);
    $('#c_branch').text(branch);
    $('#c_currency').text(currency);
    $('#totalcredits').text(total_credits);
    $('#totaldebits').text(total_debits);
}

function clearCustomerDetails() {
    $('#customer_name').text("");
    $('#c_branch').text("");
    $('#c_currency').text("");
    $('#totalcredits').text("");
    $('#totaldebits').text("");
}

function clearErrorTexts(){
    $('#acc_err').html('');
    $('#start_err').html('');
    $('#end_err').html('');
}

const displayAccountDoesNotExist = ()=>{
    //table responsive
    let tableResponsiveDiv = document.createElement('div');
    tableResponsiveDiv.classList.add('table-responsive');

    //table
    let validateTable = document.createElement('table');
    validateTable.classList.add('table')
    validateTable.classList.add('table-hover')
    validateTable.classList.add('table-bordered');

    //table head
    let tableHead = document.createElement('thead')
    tableHead.classList.add('thead-dark');
    let tHeadRow = document.createElement('tr');

    //table body
    let tableBody = document.createElement('tbody');

    let headCols = ['Entry Value', 'Value Date', 'Particulars', 'Withdrawal', 'Deposit', 'Balance'];
    //add th to tr
    for (let i = 0; i < headCols.length; i++) {

        let th = document.createElement('th');
        $(th).addClass("text-center")
        th.setAttribute('scope', 'col');
        th.innerHTML = headCols[i];

        tHeadRow.appendChild(th);
    }

    let tBodyRow = document.createElement('tr');
    $(tBodyRow).html(`<td colspan="6" class="text-danger text-center fs-5">Account Does Not Exists!! </td>`);
    tableBody.append(tBodyRow);

    $(tableHead).html(tHeadRow);
    $(validateTable).html(tableHead);
    $(validateTable).append(tableBody);
    $(tableResponsiveDiv).html(validateTable);

    return tableResponsiveDiv;
}

const displayMessageInTable = (message)=>{
    //table responsive
    let tableResponsiveDiv = document.createElement('div');
    tableResponsiveDiv.classList.add('table-responsive');

    //table
    let validateTable = document.createElement('table');
    validateTable.classList.add('table')
    validateTable.classList.add('table-hover')
    validateTable.classList.add('table-bordered');

    //table head
    let tableHead = document.createElement('thead')
    tableHead.classList.add('thead-dark');
    let tHeadRow = document.createElement('tr');

    //table body
    let tableBody = document.createElement('tbody');

    let headCols = ['Entry Value', 'Value Date', 'Particulars', 'Withdrawal', 'Deposit', 'Balance'];
    //add th to tr
    for (let i = 0; i < headCols.length; i++) {

        let th = document.createElement('th');
        $(th).addClass("text-center")
        th.setAttribute('scope', 'col');
        th.innerHTML = headCols[i];

        tHeadRow.appendChild(th);
    }

    let tBodyRow = document.createElement('tr');
    $(tBodyRow).html(`<td colspan="6" class="text-danger text-center fs-5">${message}</td>`);
    tableBody.append(tBodyRow);

    $(tableHead).html(tHeadRow);
    $(validateTable).html(tableHead);
    $(validateTable).append(tableBody);
    $(tableResponsiveDiv).html(validateTable);

    return tableResponsiveDiv;
}