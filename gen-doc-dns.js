const dnsList = require('./client/src/utils/dns-list.json');

console.log(dnsList);

// for each make a request to https://go-acme.github.io/lego/dns/{dns}/#credentials

let finalList = {};

const fs = require('fs');

let i = 0;
(async () => {
  let resultList = (await (await fetch(`https://go-acme.github.io/lego/dns/`)).text());
  resultList = resultList.split(`<h2 id="dns-providers">DNS Providers</h2>`)[1];
  resultList = resultList.split(`</table>`)[0] + `</table>`;

  let dnses = resultList.match(/<code>(.*?)<\/code>/g);
  dnses = dnses.map(v => v.replace(/<\/?code>/g, ''));

  // remove exec
  dnses = dnses.filter(v => v !== 'exec' && v !== 'hyperone' && v !== 'manual');

  fs.writeFileSync('./client/src/utils/dns-list.json', JSON.stringify(dnses, null, 2));

  for(const dns of dnsList) {
    console.log(`Fetching ${dns} infos`)
    let result = (await (await fetch(`https://go-acme.github.io/lego/dns/${dns}/#credentials`)).text());
    result = result.split(`<h2 id="credentials">Credentials</h2>`)[1];
    let result2 = result.split(`<h2 id="additional-configuration">Additional Configuration</h2>`)[1];
    result = result.split(`</table>`)[0] + `</table>`;
    let vars = result.match(/<code>(.*?)<\/code>/g);
    vars = vars.map(v => v.replace(/<\/?code>/g, ''));

    // additional vars
    if(result2) {
      result2 = result2.split(`</table>`)[0] + `</table>`;
      let vars2 = result2.match(/<code>(.*?)<\/code>/g);
      vars2 = vars2.map(v => v.replace(/<\/?code>/g, ''));
      vars = vars.concat(vars2);
    }

    // filter out non env-var (AZaz09_)
    vars = vars.filter(v => v.match(/^[A-Z_][A-Z0-9_]*$/));

    finalList[dns] = {
      name: dns,
      url: `https://go-acme.github.io/lego/dns/${dns}/#credentials`,
      docs: result + result2,
      vars: vars
    }

    console.log(`${i++}/${dnsList.length} done`)
  }

  // save to file
  fs.writeFileSync('./client/src/utils/dns-config.json', JSON.stringify(finalList, null, 2));
})();