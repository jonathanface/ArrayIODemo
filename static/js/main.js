/*jshint esversion: 6 */

var availableUsers = [{
  firstName: 'Donald',
  lastName: 'Blair',
  ssn: '666285344',
  dob: '1939-09-20',
  address: {
    street: '3627 W Poplar St',
    city: 'San Antonio',
    state: 'TX',
    zip: '78228'
  },
  auth:false
},
{
  "firstName":"Maria",
  "lastName":"Iglesias",
  "address": { 
    "street":"21 Pacific St",
    "city":"Pittsfield",
    "state":"MA",
    "zip":"01201"
  },
  "dob":"1958-12-28",
  "ssn":"666824123"
},
{
  "firstName":"Thomas",
  "lastName":"Friedman",
  "address": {
    "street":"535 30 RD A",
    "city":"Grand Junction",
    "state":"CO",
    "zip":"81504"
  },
  "dob":"1975-01-01",
  "ssn":"666234390"
},{
    "firstName":"Kimberly",
    "lastName":"Olstrup",
    "address": {
        "street":"125 W 100 N",
        "city":"Jerome",
        "state":"ID",
        "zip":"83338",
    },
    "dob":"1947-01-01",
    "ssn":"666561858"
}];
var addedUsersList = new Map();
var currentSSN;

function clearOrStartStandby(on) {
  if (!on) {
    document.querySelector('.standby').style = 'display:none';
  } else {
    document.querySelector('.standby').style = 'display:initial';
  }
}

function updateSelectBox() {
  let select = document.querySelector('#availableUsers');
  select.innerHTML = '';
  let opt = document.createElement('option');
  opt.value = null;
  opt.innerHTML = 'Select a user...';
  select.appendChild(opt);
  for (let i=0; i < availableUsers.length; i++) {
    let opt = document.createElement('option');
    opt.value = availableUsers[i].ssn;
    opt.innerHTML = availableUsers[i].lastName + ", " + availableUsers[i].firstName;
    select.appendChild(opt);
  }
}

window.onload = function() {
  updateSelectBox();
};

function getCreditReport(reportKey, displayToken) {
  
  clearOrStartStandby(true);
  let url = new URL('/api/report/view', window.location.origin);
  let json = {'reportKey': reportKey, 'displayToken':displayToken};
  url.search = new URLSearchParams(json).toString();
  fetch(url, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    }
  }).then(data => {
    let json;
    
    switch(data.status) {
      case 202:
        setTimeout(function() {
          getCreditReport(reportKey, displayToken);
        }, 1000);
        break;
      case 200:
        clearOrStartStandby();
        json = data.json().then(json => {
          var jsonPretty = JSON.stringify(json, null, 4); 
          document.querySelector('#creditData').innerHTML = jsonPretty;
          document.querySelector('#userFormHolder').style = 'display:none';
          document.querySelector('#creditReportHolder').style = 'display:initial';
        });
        break;
      default:
        console.log('error: ' + data.status);
        clearOrStartStandby();
    }
      
  });
}

function goBack() {
  document.querySelector('#userFormHolder').style = 'display:initial';
  document.querySelector('#creditReportHolder').style = 'display:none';
}

function revokeAuth() {
  clearOrStartStandby(true);
  let json = {'id': currentSSN};
  fetch('/api/user/revoke', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(json)
  })
  .then(data => {
    let json;
    clearOrStartStandby();
    switch(data.status) {
      case 200:
        updateAddedUsersList();
        break;
      default:
        console.log('error: ' + data.status);
    }
  });
}

function requestCreditReport() {
  document.querySelector('.error').innerHTML = '';
  let json = {'id': currentSSN};
  clearOrStartStandby(true);
  fetch('/api/report/request', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(json)
  })
  .then(data => {
    let json;
    
    switch(data.status) {
      case 200:
        json = data.json().then(json => {
          getCreditReport(json.reportKey, json.displayToken);
        });
        break;
      case 202:
        json = data.json().then(json => {
          console.log(json);
        });
        break;
      default:
        console.log('error: ' + data.status);
        clearOrStartStandby();
    }
  });
}

function onQuestionFormEnter() {
  console.log('submitting...');
  clearOrStartStandby(true);
  document.querySelector('.error').innerHTML = '';
  document.querySelector('#questionFormHolder .error').innerHTML = '';
  let answers = new Map();
  let buttons = document.querySelectorAll('input[type="radio"]');
  let qCount = document.querySelectorAll('p[data-id').length;
  let answered = 0;
  buttons.forEach(button => {
    if (button.checked) {
      answered++;
      answers[(button.getAttribute('name')).toString()] = button.dataset.id.toString();
    }
  });
  if (answered < qCount) {
    document.querySelector('#questionFormHolder .error').innerHTML = 'You must answer all questions.';
    return;
  }
  let json = {};
  json.id = currentSSN;
  json.answers = answers;
  
  console.log('json', json);
  fetch('/api/qa', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(json)
  })
  .then(data => {
    let json;
    console.log('stat', data.status);
    clearOrStartStandby();
    switch(data.status) {
      case 200:
        addedUsersList[currentSSN].auth = true;
        document.querySelector('#questionFormHolder').style="display:none";
        document.querySelector('#userFormHolder').style='display:initial';
        updateAddedUsersList();
        break;
      case 206:
        json = data.json().then(json => {
          console.log('success', json);
          renderQuestions(json);
        });
        break;
      default:
        console.log('fail', data);
        
        document.querySelector('#userFormHolder').style.display = 'initial';
        document.querySelector('.error').innerHTML = 'Identity validation failed';
        document.querySelector('#questionFormHolder').style.display = 'none';
    }
  });
}

function renderQuestions(json) {
  document.querySelector('#questionsHolder').innerHTML = '';
  json.questions.forEach(question => {
    console.log('question', question);
    let p = document.createElement('p');
    p.setAttribute('data-id', question.id);
    let h3 = document.createElement('h4');
    h3.innerHTML = question.text;
    p.appendChild(h3);
    let ul = document.createElement('ul');
    question.answers.forEach(answer => {
      let li = document.createElement('li');
      
      let label = document.createElement('label');
      let input = document.createElement('input');
      input.setAttribute('type', 'radio');
      input.setAttribute('name', question.id);
      input.setAttribute('data-id', answer.id);
      label.appendChild(input);
      label.appendChild(document.createTextNode(answer.text));
      li.appendChild(label);
      ul.appendChild(li);
    });
    p.appendChild(ul);
    document.querySelector('#questionsHolder').appendChild(p);
  });
  document.querySelector('#questionFormHolder').style="display:initial";
  
}

function requestQuestions() {
  console.log('getting questions for... '  + currentSSN);
  clearOrStartStandby(true);
  let url = new URL('/api/user/questions', window.location.origin);
  let params = {id:currentSSN};
  url.search = new URLSearchParams(params).toString();
  fetch(url, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    }
  })
  .then(data => {
    let json;
    console.log('stat', data.status);
    clearOrStartStandby();
    switch(data.status) {
      case 200:
        json = data.json().then(json => {
          document.querySelector('#userFormHolder').style="display:none";
          renderQuestions(json);
        });
        break;
      default:
        console.log('code', data.status);
        json = data.json().then(json => {
          if (json.error) {
            json.error.forEach(error => {
              document.querySelector('#' + cleanString(error.param)).innerHTML = error.msg;
            });
          }
        });
    }
  });
}



function updateAddedUsersList() {
  document.querySelector('#addedUsers').innerHTML = '';
  console.log('update added users', addedUsersList);
  for (var key in addedUsersList) {
    let li = document.createElement('li');
    let btn = document.createElement('button');
    console.log('add', addedUsersList[key].auth);
    if (addedUsersList[key].auth) {
      btn.innerHTML = "Revoke";
      btn.title = 'Revoke authentication';
      btn.onclick = function () {
        currentSSN = null;
        addedUsersList[key].auth = false;
        revokeAuth();
      };
    } else {
      btn.title = 'Answer financial questions to verify identitiy';
      btn.innerHTML = "Authenticate";
      btn.onclick = function () {
        currentSSN = key;
        requestQuestions();
      };
    }
    
    li.appendChild(document.createTextNode(addedUsersList[key].lastName + ', ' + addedUsersList[key].firstName));
    li.appendChild(btn);
    btn = document.createElement('button');
    btn.innerHTML = 'Credit Report';
    if (!addedUsersList[key].auth) {
      btn.disabled = true;
      btn.title = 'User must be authenticated';
    }
    btn.onclick = function() {
      currentSSN = key;
      requestCreditReport();
    };
    li.appendChild(btn);
    
    btn = document.createElement('button');
    btn.innerHTML = 'Delete';
    btn.title = 'Delete user from the system';
    btn.onclick = function() {
      if (addedUsersList[key].auth) {
        addedUsersList[key].auth = false;
        revokeAuth();
      }
      currentSSN = null;
      availableUsers.push(addedUsersList[key]);
      delete addedUsersList[key];
      updateAddedUsersList();
      updateSelectBox();
    };
    li.appendChild(btn);
    document.querySelector('#addedUsers').appendChild(li);
  }
}

function onUserFormEnter(event) {
  let select = document.querySelector('#availableUsers');
  let userID = select.options[select.selectedIndex].value;
  if (!userID) {return;}
  let user;
  for (let i=0; i < availableUsers.length; i++) {
    if (availableUsers[i].ssn == userID) {
      user = availableUsers[i];
      break;
    }
  }
  if (!user) {return;}
  clearOrStartStandby(true);
  fetch('/api/user', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(user)
  })
  .then(data => {
    let json;
    clearOrStartStandby();
    switch(data.status) {
      case 201:
        availableUsers.splice(availableUsers.indexOf(user),1);
        addedUsersList[user.ssn] = user;
        updateAddedUsersList();
        updateSelectBox();
        break;
      default:
        console.log('code', data.status);
        json = data.json().then(json => {
          if (json.error) {
            json.error.forEach(error => {
              document.querySelector('.error').innerHTML = error.msg;
            });
          }
        });
    }
  });
}