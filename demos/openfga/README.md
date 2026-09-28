# OpenFGA demo

This folder contains the presentation and data used for the OpenFGA demo.

## Run the OpenFGA playground

Run the following command:
```shell
docker run --rm -p 8080:8080 -p 3000:3000 openfga/openfga run
```

Now open your browser on http://localhost:3000/playground.

When you are finished with the playground, go back to the command-line and just press <ctrl-c> to stop OpenFGA.

## Example models & tuples

To load the examples used during the presentation, run the following command while the playground is active:
```shell
cd demos/openfga
./all-examples.sh
cd -
```

Please refresh your browser window if you already opened the playground.

OpenFGA keeps its data in memory, so stopping it deletes all data.
