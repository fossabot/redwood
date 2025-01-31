const Redwood = require('../../redwood.js').default
const fs = require('fs')

//
// Redwood setup
//
let node1Identity = Redwood.identity.random()
let node1Client = Redwood.createPeer({
    identity: node1Identity,
    httpHost: 'http://localhost:8080',
    onFoundPeersCallback: (peers) => {}
})

let node2Identity = Redwood.identity.random()
let node2Client = Redwood.createPeer({
    identity: node2Identity,
    httpHost: 'http://localhost:9090',
    onFoundPeersCallback: (peers) => {}
})

async function main() {
    await node1Client.authorize()
    await node2Client.authorize()
    await genesis()
    console.log('Done.')
    process.exit(0)
}

async function genesis() {
    // Upload the repo's files into the state tree.
    let indexHTML = fs.createReadStream('./repo/index.html')
    let scriptJS = fs.createReadStream('./repo/script.js')
    let readmeMD = fs.createReadStream('./repo/README.md')
    let redwoodJPG = fs.createReadStream('./repo/redwood.jpg')
    let { sha1: indexHTMLSha1 } = await node1Client.storeRef(indexHTML)
    let { sha1: scriptJSSha1 } = await node1Client.storeRef(scriptJS)
    let { sha1: readmeMDSha1 } = await node1Client.storeRef(readmeMD)
    let { sha1: redwoodJPGSha1 } = await node1Client.storeRef(redwoodJPG)

    // Send the genesis tx to set up the repo (this is sort of like `git init`).
    //
    // The "somegitprovider.org/gitdemo" channel contains:
    //   - A link to the current worktree so that we can browse it like a regular website.
    //   - All of the commit data that Git expects.  The files are stored under a "files" key,
    //         and other important metadata are stored under the other keys.
    //   - A mapping of refs (usually, branches) to commit hashes.
    //   - A permissions validator that allows anyone to write to the repo but tries to keep
    //         people from writing to the wrong keys.
    let tx1 = {
        stateURI: 'somegitprovider.org/gitdemo',
        id: Redwood.utils.genesisTxID,
        parents: [],
        patches: [
            ' = ' + Redwood.utils.JSON.stringify({
                'demo': {
                    'Content-Type': 'link',
                    'value': 'state:somegitprovider.org/gitdemo/refs/heads/master/worktree'
                },
                'Merge-Type': {
                    'Content-Type': 'resolver/dumb',
                    'value': {}
                },
                'Validator': {
                    'Content-Type': 'validator/permissions',
                    'value': {
                        [node1Identity.address.toLowerCase()]: {
                            '^.*$': {
                                'write': true
                            }
                        },
                        '*': {
                            '^\\.refs\\..*': {
                                'write': true
                            },
                            '^\\.commits\\.[a-f0-9]+\\.parents': {
                                'write': true
                            },
                            '^\\.commits\\.[a-f0-9]+\\.message': {
                                'write': true
                            },
                            '^\\.commits\\.[a-f0-9]+\\.timestamp': {
                                'write': true
                            },
                            '^\\.commits\\.[a-f0-9]+\\.author': {
                                'write': true
                            },
                            '^\\.commits\\.[a-f0-9]+\\.committer': {
                                'write': true
                            },
                            '^\\.commits\\.[a-f0-9]+\\.files': {
                                'write': true
                            }
                        }
                    }
                },
                'refs': {
                    'heads': {}
                },
                'commits': {}
            }),
        ],
    }
    await node1Client.put(tx1)

    //
    // Now, let's simulate pushing the first commit.  This transaction looks exactly like the ones
    // generated by the Redwood git-remote-helper.
    //

    // Note: if you alter the contents of the ./repo subdirectory, you'll need to determine the
    // git commit hash of the first commit again, and then tweak these variables.  Otherwise,
    // you'll get a "bad object" error from git.
    let commit1Hash = '3832ec27a841e7a2071e1edaa6ca280f974f20b3'
    let commit1Timestamp = '2021-03-01T18:03:22-06:00'

    let tx2 = {
        stateURI: 'somegitprovider.org/gitdemo',
        id: commit1Hash + '000000000000000000000000',
        parents: [ tx1.id ],
        patches: [
            `.commits.${commit1Hash} = ` + Redwood.utils.JSON.stringify({
                'message': 'First commit\n',
                'timestamp': commit1Timestamp,
                'author': {
                    'name': 'Bryn Bellomy',
                    'email': 'bryn.bellomy@gmail.com',
                    'timestamp': commit1Timestamp,
                },
                'committer': {
                    'name': 'Bryn Bellomy',
                    'email': 'bryn.bellomy@gmail.com',
                    'timestamp': commit1Timestamp,
                },
                'files': {
                    'README.md': {
                        'Content-Type': 'link',
                        'mode': 33188,
                        'value': `ref:sha1:${readmeMDSha1}`,
                    },
                    'redwood.jpg': {
                        'Content-Type': 'link',
                        'mode': 33188,
                        'value': `ref:sha1:${redwoodJPGSha1}`,
                    },
                    'index.html': {
                        'Content-Type': 'link',
                        'mode': 33188,
                        'value': `ref:sha1:${indexHTMLSha1}`,
                    },
                    'script.js': {
                        'Content-Type': 'link',
                        'mode': 33188,
                        'value': `ref:sha1:${scriptJSSha1}`,
                    }
                }
            }),
            `.refs.heads.master = ` + Redwood.utils.JSON.stringify({
                'HEAD': commit1Hash,
                'worktree': {
                    'Content-Type': 'link',
                    'value': `state:somegitprovider.org/gitdemo/commits/${commit1Hash}/files`
                }
            }),
        ],
    }
    await node1Client.put(tx2)
}

main()
