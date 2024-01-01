#include <typeinfo>
#include <fstream>
#include "emp-sh2pc/emp-sh2pc.h"
#include "plain_biometric.cpp"
#include "libs/CLI11.hpp"

using namespace emp;
using namespace std;

const bool VERBOSE = false;
int *secret, *alice_share, *bob_share;

const int MAX_SIMILARITY_SCORE = 4024000;
const int SIMILARITY_THRESHOLD = 10000;
const int PRIME_MOD = 0x3ee0001;

void hyb_janus_threshold(int party, int *secret_shares, BioSetting bio) {
    // set public parameters
    const int BITLEN = 28;
	Integer T(BITLEN, bio.threshold, PUBLIC);
	Integer P(BITLEN, bio.prime_mod, PUBLIC);

    int size = bio.db_size;

    // Set input
	Integer *S1 = new Integer[size];
	Integer *S2 = new Integer[size];
	Integer *S = new Integer[size];
	Bit *Match = new Bit[size];
	Bit Result = new Bit(false, ALICE);


    for (int i = 0; i < size; i++)
        S1[i] = Integer(BITLEN, secret_shares[i], ALICE);
    for (int i = 0; i < size; i++)
        S2[i] = Integer(BITLEN, secret_shares[i], BOB);
    cerr << "  # Setting input finished." << endl;

    // Compute threshold
    for (int i = 0; i < size; i++){
        S[i] = (S1[i] + S2[i]);
        
        Bit overflow = S[i] >= P;
        Integer correction = Integer(BITLEN, 0, PUBLIC).select(overflow, P);
        S[i] = S[i] - correction;

        // fingercode
        if (bio.bio_type == "finger")
            Match[i] = S[i] < T;
        // Optimization: threshold using MSBs instead of < for our iris code setting
        else if (bio.bio_type == "iris")
            Match[i] = S[i].bits[BITLEN-1] | S[i].bits[BITLEN-2];
    }
    cerr << "  # Computing matches finished." << endl;
    
    // Fuse matches of f templates for each user and compute the binary membership result
    for (int i = 0; i < bio.user_num; i++){
        Bit fmatch  = Match[i*bio.fuse];
        for (int j = 1; j < bio.fuse; j++){
            fmatch = fmatch & Match[i*bio.fuse+j];
        }
        Result = Result | fmatch;
    }

    cout << "Membership result: " << Result.reveal<bool>(ALICE)  << endl;
    cerr << "  # Revealing matches finished." << endl;
}

int positive_mod(int a, int p){
    return (a % p + p) % p;
}


// generate fake data for benchmarking
int* gen_fake_fhe_data(int party, BioSetting bio, int max_secret){
    PRG prg(fix_key);

    int size = bio.db_size;
    secret = new int[size];
    alice_share = new int[size];
    bob_share = new int[size];

    prg.random_data(secret, size*sizeof(secret[0]));
    prg.random_data(alice_share, size*sizeof(alice_share[0]));

	for(int i = 0; i < size; ++i) {
        secret[i] = positive_mod(secret[i], max_secret);
        alice_share[i] = positive_mod(alice_share[i], bio.prime_mod);
        bob_share[i] = positive_mod(secret[i] - alice_share[i], bio.prime_mod);
    }

    if(VERBOSE){
        for(int i = 0; i < size; ++i) {
            cerr << "sec[" << i << "]: " << secret[i] << " a:" << alice_share[i] << " - b: " 
                << bob_share[i] <<  " -> " << (alice_share[i] + bob_share[i])%bio.prime_mod << endl; 
        }
    }

    if (party == ALICE)
        return alice_share;
    else
        return bob_share;
}


int main(int argc, char** argv) {
	string party_str;           // party from  [RS, BP]
    int port;                   // network port
    string addr;                // address of the benchmark file
    string bio_type = "finger"; // biometric type (we optimize comparison for iris to only or MSBs)
    int N = 32;                 // number of users
    int fuse = 1;               // number of templates per user


    CLI::App app{"Hyb-Janus threshold component"};
    app.add_option("party", party_str, "Party: [RS, BP]")->required();
    app.add_option("port", port, "Network port")->required();
    app.add_option("--bio-type", bio_type, "Biometric type: [finger, iris]");
    app.add_option("--addr", addr, "Address of the output benchmark file");
    app.add_option("-N", N, "Number of registered users (N)");
    app.add_option("-f", fuse, "Number of biometric templates per user (f)");

    try {
        app.parse(argc, argv);
    } catch (const CLI::ParseError &e) {
        return app.exit(e);
    }

    int party = 0;
    if (party_str == "RS")
        party = ALICE;
    else if (party_str == "BP")
        party = BOB;
    else {
        cout << "Invalid party: " << party_str << endl;
        return 1;
    }   

    ofstream fout;
    fout.open(addr, fout.out | fout.app);

    if (VERBOSE){
        cerr << "party: " << party_str << " (" << party <<  ") port: " << port << " bio-type: " <<  bio_type << " addr: " << addr << " N: " << N << " fuse: " << fuse << endl;
    }

	NetIO * io = new NetIO(party==ALICE ? nullptr : "127.0.0.1", port);
	setup_semi_honest(io, party);


    // template size is not affecting the threshold computation, so ts is set to 0 here
    BioSetting bio(bio_type, N, fuse, 0, SIMILARITY_THRESHOLD, PRIME_MOD);

    // Instead of using real data from running the SHE portion and secret sharing its output, we generate fake data as we are only interested in the performance of the threshold computation.
    auto secret_shares = gen_fake_fhe_data(party, bio, MAX_SIMILARITY_SCORE);

    auto start = chrono::high_resolution_clock::now();
    hyb_janus_threshold(party, secret_shares, bio); 
    auto end = chrono::high_resolution_clock::now();
    auto duration = chrono::duration_cast<chrono::milliseconds>(end-start).count();

    cerr << "$$$$$ Execution finished successfully $$$$$$" << endl;

    auto transfer_byte = io->get_sent_byte_count();

    // report cost
    cout << "Execution of party " << party_str << "finished successfuly.\n" <<
            "    Duration: " << duration << " ms\n" << 
            "    Bytes sent by " << party_str << ": " << transfer_byte << "B. ("  <<
            transfer_byte/1024/1024 << " MB)" <<  endl;

    // write execution detail to the benchmark file
    fout << N << ", " << fuse << ", " << duration << ", " << transfer_byte << endl;

	finalize_semi_honest();
	delete io;
    return 0;
}
