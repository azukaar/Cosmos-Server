const fs = require('fs').promises;
const path = require('path');
const { execSync } = require('child_process');
const axios = require('axios');

const API_KEY = process.env.API_KEY;
const BATCH_SIZE = 50;

async function readJsonFile(filePath) {
  const data = await fs.readFile(filePath, 'utf8');
  return JSON.parse(data);
}

async function writeJsonFile(filePath, data) {
  await fs.writeFile(filePath, JSON.stringify(data, null, 2));
}

async function translateBatch(batch, targetLang) {
  const prompt = `Translate the following JSON key-value pairs from English to ${targetLang}. Maintain the JSON structure and only translate the values. Always return valid JSON:

${JSON.stringify(batch, null, 2)}`;

  try {
    const res = await fetch('https://api.openai.com/v1/chat/completions', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${API_KEY}`,
        },
        body: JSON.stringify({
            model: "gpt-4o",
            max_tokens: 3000,
            messages: [
                {
                    role: "system",
                    content: "You are a translator for technical products.",
                },
                {
                    role: "user",
                    content: prompt,
                },
            ],
        })
    });

    let rawText = await res.json();
    // console.log(rawText);
    rawText = rawText.choices[0].message.content;
    let json = rawText.split('{').slice(1).join('{');
    json = json.split('}').slice(0, -1).join('}');
    console.log(json);
    json = JSON.parse(`{${json}}`);

    return json;
  } catch (error) {
    console.error('Error calling API:', error.message);
    return null;
  }
}

async function getModifiedKeys(filePath, since) {
  try {
    const output = execSync(`git diff ${since} -- ${filePath}`, { encoding: 'utf-8' });
    const addedLines = output.split('\n').filter(line => line.startsWith('+') && line.includes(':'));
    return addedLines.map(line => line.split(':')[0].replace(/["+]/g, '').trim());
  } catch (error) {
    console.error('Error getting modified keys:', error.message);
    return [];
  }
}

async function translateFile(sourcePath, targetPath, targetLang, all, since) {
  const sourceData = await readJsonFile(sourcePath);
  let targetData = {};

  if (!all) {
    try {
      targetData = await readJsonFile(targetPath);
    } catch (error) {
      console.log('Target file not found. Creating a new one.');

      // Create the target directory if it doesn't exist
      const targetDir = path.dirname(targetPath);
      await fs.mkdir(targetDir, { recursive: true });
    }
  }

  let keysToTranslate;
  if (all) {
    keysToTranslate = Object.keys(sourceData);
  } else if (since) {
    keysToTranslate = await getModifiedKeys(sourcePath, since);
  } else {
    keysToTranslate = Object.keys(sourceData).filter(key => !(key in targetData));
  }

  for (let i = 0; i < keysToTranslate.length; i += BATCH_SIZE) {
    const batch = Object.fromEntries(
      keysToTranslate.slice(i, i + BATCH_SIZE).map(key => [key, sourceData[key]])
    );
    
    console.log(`Translating keys ${i + 1} to ${i + BATCH_SIZE} out of ${keysToTranslate.length}`);

    const translatedBatch = await translateBatch(batch, targetLang);
    if (translatedBatch) {
      Object.assign(targetData, translatedBatch);
    }
  }

  if (!all) {
    // Remove keys that are not in the source file
    Object.keys(targetData).forEach(key => {
      if (!(key in sourceData)) {
        delete targetData[key];
      }
    });
  }

  await writeJsonFile(targetPath, targetData);
  console.log(`Translation completed. Output saved to ${targetPath}`);
}

async function main() {
  const args = process.argv.slice(2);
  const targetLang = args[0];
  const all = args.includes('--all');
  const sinceIndex = args.indexOf('--since');
  const since = sinceIndex !== -1 ? args[sinceIndex + 1] : null;

  if (!targetLang) {
    console.error('Please provide a target language.');
    process.exit(1);
  }

  if (all && since) {
    console.error('--all and --since options are incompatible.');
    process.exit(1);
  }

  const sourcePath = path.join('client/src/utils/locales', 'en', 'translation.json');
  const targetPath = path.join('client/src/utils/locales', targetLang, 'translation.json');

  await translateFile(sourcePath, targetPath, targetLang, all, since);
}

main().catch(console.error);