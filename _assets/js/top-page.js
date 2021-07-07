'use strict';
const main = () => {
  let data;
  let count = 51; // default

  const fetch = () => {
    const file = '/data/post-data.json';
    const xhr = new XMLHttpRequest();
    xhr.onreadystatechange = () => {
      if (xhr.readyState === 4 && xhr.status === 200) {
        if (xhr.response) {
           data = JSON.parse(xhr.responseText);
        }
      }
    }
    xhr.open('GET', file, true);
    xhr.setRequestHeader('Cache-Control', 'no-cache');
    xhr.send(null);
  }

  const insertContents = (obj, start, end) => {
    let html = [];
    for (let i = start; i < end; i++) {
      html.push(`
        <li class='related-list-item'>
          <a href='/${obj[i].slug}/' class='related-link'>
            <div class="content">
              <div class='item-title'>${obj[i].title}</div>
              <div class=${obj[i].thumnail ? 'thumnail' : 'dscr'}>${obj[i].summary}</div>
            </div>
          </a>
        </li>
        `);
    } 
    document.getElementById('link-list').insertAdjacentHTML('beforeEnd', html.join(''));
  }

  const more = (e) => {
    const plen = btn.dataset.len;
    const amount = plen - count > count ? count : plen - count; 
    const end = count + amount;
    insertContents(data, count, end);
    count += amount;
    if (plen - end <= 0) {
      btn.parentNode.removeChild(btn);
    }
  }
  const btn = document.getElementById('more');
  if (btn) {
    btn.addEventListener('click', more, false);
    count = Number(btn.dataset.limit);
    fetch();
  }
}
document.addEventListener('DOMContentLoaded', main);
