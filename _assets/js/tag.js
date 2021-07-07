'use strict';
const main = () => {
  let data;
  let count = 50; // default

  const fetch = () => {
    const file = '/data/tag-data.json';
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

  const insertContents = (obj, slug, start, end) => {
    let html = [];
    const tObj = obj[slug];
    for (let i = start; i < end; i++) {
      html.push(`
        <li class='related-list-item'>
          <a href='/${tObj[i].slug}/' class='related-link'>
            <div class="content">
              <div class='item-title'>${tObj[i].title}</div>
              <div class=${tObj[i].thumnail ? 'thumnail' : 'dscr'}>${tObj[i].summary}</div>
            </div>
          </a>
        </li>
        `);
    } 
    document.getElementById('link-list').insertAdjacentHTML('beforeEnd', html.join(''));
  }

  const more = (e) => {
    const plen = btn.dataset.len;

    const slug = btn.dataset.slug;
    const amount = plen - count > count ? count : plen - count; 
    const end = count + amount;
    insertContents(data, slug, count, end);
    count += amount;
    if (plen - end <= 0) {
      btn.parentNode.removeChild(btn);
    }
  }

  const btn = document.getElementById('more');
  if (btn) {
    btn.addEventListener('click', more, false);
    count =  Number(btn.dataset.limit);
    fetch();
  } 
}

document.addEventListener('DOMContentLoaded', main);

