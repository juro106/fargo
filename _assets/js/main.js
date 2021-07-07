'use strict';
const main = () => {
  let data;
  const fetch = () => {
    const xhr = new XMLHttpRequest();
    const file = '/data/tag-data.json';
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

  const insertContents = async (obj, slug, start) => {
    return new Promise(resolve => {
      let html = [];
      const tObj = obj[slug];
      const len = tObj.length;
      for (let i = start; i < len; i++) {
        html.push(`
          <li class='related-list-item'>
            <a class='related-link' href='/${tObj[i].slug}/'>
              <div class="content">
                <div class='item-title'>${tObj[i].title}</div>
                <div class=${tObj[i].thumnail ? 'thumnail' : 'dscr'}>${tObj[i].summary}</div>
              </div>
            </a>
          </li>
        `);
      } 
      document.getElementById(`ulid_${slug}`).insertAdjacentHTML('beforeEnd', html.join(''));
      resolve();
    });
  }

  const remove = async (target) => {
    return new Promise(resolve => {
      target.parentNode.removeChild(target);
    });
  }

  const loading = (target, data) => {
    return new Promise((resolve, reject) => {
        target.innerHTML = "<div class='spinner'></div>";
        resolve(data);
    });
  }

  const wait = (target) => {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        resolve();
      }, 0);
    });
  }

  const more = async (e) => {
    const target = e.currentTarget;
    const slug = target.dataset.slug;
    const start = Number(target.dataset.limit);
    loading(target);
    await wait(target);

    // const data = await fetch();
    await insertContents(data, slug, start);
    await remove(target);
  }
  const mores = document.querySelectorAll('[data-class="more"]'); 
  if (mores.length > 0) {
    mores.forEach(v => v.addEventListener('click', more, false));
    fetch();
  }
}

document.addEventListener('DOMContentLoaded', main);

