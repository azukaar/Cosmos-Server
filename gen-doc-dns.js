// for each make a request to https://go-acme.github.io/lego/dns/{dns}/#credentials

let finalList = {};

const fs = require('fs');

// headings carry a trailing anchor <span>, so split on the opening tag then drop the rest of the heading
const afterHeading = (html, id) => {
  const rest = html.split(new RegExp(`<h2 id="${id}"`))[1];
  if (!rest) return undefined;
  const end = rest.indexOf(`</h2>`);
  return end === -1 ? rest : rest.slice(end + 5);
};

let i = 0;
(async () => {
  let resultList = (await (await fetch(`https://go-acme.github.io/lego/dns/`)).text());
  resultList = afterHeading(resultList, 'dns-providers');
  resultList = resultList.split(`</table>`)[0] + `</table>`;

  let dnses = resultList.match(/<code>(.*?)<\/code>/g);
  dnses = dnses.map(v => v.replace(/<\/?code>/g, ''));

  // remove exec
  dnses = dnses.filter(v => v !== 'exec' && v !== 'hyperone' && v !== 'manual');

  fs.writeFileSync('./client/src/utils/dns-list.json', JSON.stringify(dnses, null, 2));

  for(const dns of dnses) {
    console.log(`Fetching ${dns} infos`)
    let result = (await (await fetch(`https://go-acme.github.io/lego/dns/${dns}/#credentials`)).text());
    result = afterHeading(result, 'credentials');
    let result2 = afterHeading(result, 'additional-configuration');
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
      docs: result + (result2 || ''),
      vars: vars
    }

    console.log(`${i++}/${dnses.length} done`)
  }

  // save to file
  fs.writeFileSync('./client/src/utils/dns-config.json', JSON.stringify(finalList, null, 2));
})();