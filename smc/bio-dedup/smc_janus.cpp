#include <typeinfo>
#include <chrono>
#include <vector>
#include "emp-sh2pc/emp-sh2pc.h"
#include "plain_biometric.cpp"
#include "libs/CLI11.hpp"
using namespace emp;
using namespace std;

const bool VERBOSE = false;
const int SIMILARITY_THRESHOLD = 5000;
PlainBiometric **plain_bio_db, **rs_bio_share_db, **bp_bio_share_db, *raw_query;


/**
 * Generate a random biometric database then secret share it between the 
 * registration station and the biometric provider.
 * This code uses deterministic randomness and runs on both the registration station 
 * and the biometric provider.
 * 
 * Our goal is evaluating the performance. In here, both parties have read access
 * to all the templates.
 * In a real setting, each party should only have its share in storage and will 
 * not have access to the raw data.
*/
void gen_data(BioSetting bio){
    cerr << "Generating raw biometric data..." << endl;

    raw_query = gen_rnd_biometric(bio.template_size, bio.bits_per_slot, 9);
    plain_bio_db = gen_rnd_bio_db(bio.db_size, bio.template_size, bio.bits_per_slot, 142);
    rs_bio_share_db = new PlainBiometric*[bio.db_size];
    bp_bio_share_db = new PlainBiometric*[bio.db_size];

    cerr << "Secret sharing raw data between RS and BP." << endl;

    for (int i = 0; i < bio.db_size; i++){
        auto ss = secret_share(plain_bio_db[i]);
        rs_bio_share_db[i] = ss.first;
        bp_bio_share_db[i] = ss.second;
    }
    cerr << "Data generated successfully." << endl;
}


void membership_hamming(int party, BioSetting bio) {
    cerr << "Party (" <<  party << ") run hamming membership with " << bio.to_str() << endl; 
    // set public parameters
	Integer T(16, SIMILARITY_THRESHOLD, PUBLIC);

    // SMC input variables
	vector<Bit> *rs_db = new vector<Bit>[bio.db_size];
	vector<Bit> *bp_db = new vector<Bit>[bio.db_size];
	vector<Bit> match(bio.db_size);
    Bit membership = Bit(false, PUBLIC);
    
    // set input
    for (int i = 0; i < bio.db_size; i++){
        rs_db[i] = vector<Bit>(bio.template_size);
        bp_db[i] = vector<Bit>(bio.template_size);
        for(int j = 0; j < bio.template_size; j++){
            rs_db[i][j] = Bit(rs_bio_share_db[i]->data[j]^raw_query->data[j], ALICE);
            bp_db[i][j] = Bit(bp_bio_share_db[i]->data[j], BOB);
        }
    }
    cerr << "  # Setting input finished." << endl;

    // Compute diistnace and threshold
    Integer distance[bio.db_size];
    for (int i = 0; i < bio.db_size; i++){
        // represent bit vector as long integers to use internal hamming_weight function
        Integer rs = Integer(rs_db[i]);
        Integer bp = Integer(bp_db[i]);
        distance[i] = rs ^ bp ;
        distance[i] = distance[i].hamming_weight().resize(16);

        match[i] = distance[i] < T;
    }

    // Fuse matches of f templates for each user and compute the binary membership result
    for (int i = 0; i < bio.user_num; i++){
        Bit fmatch = match[i*bio.fuse];
        for (int j = 0; j < bio.fuse; j++)
            fmatch = fmatch & match[i*bio.fuse + j];
        membership = membership | fmatch;
    }

    cout << "   Membership: " << membership.reveal<bool>(ALICE) << endl;
    cerr << "  # Revealing matches finished." << endl;
}


void membership_masked_hamming(int party, BioSetting bio) {
    cerr << "Party (" <<  party << ") run masked hamming membership with " << bio.to_str() << endl; 
    // set public parameters threshold =  0.4 = 52/128
	Integer T(16+7, 52, PUBLIC);

    // create SMC input variables
	vector<Bit> *rs_db = new vector<Bit>[bio.db_size];
	vector<Bit> *bp_db = new vector<Bit>[bio.db_size];
	vector<Bit> *rs_mask_db = new vector<Bit>[bio.db_size];
	vector<Bit> *bp_mask_db = new vector<Bit>[bio.db_size];
	vector<Bit> query_mask_bits(bio.template_size);
	vector<Bit> match(bio.db_size);
    Bit membership = Bit(false, PUBLIC);
    
    // set input values
    for (int i = 0; i < bio.db_size; i++){
        rs_db[i] = vector<Bit>(bio.template_size);
        bp_db[i] = vector<Bit>(bio.template_size);
        rs_mask_db[i] = vector<Bit>(bio.template_size);
        bp_mask_db[i] = vector<Bit>(bio.template_size);
        for(int j = 0; j < bio.template_size; j++){
            // apply ^ to the query in the plain domain.
            // The mask requires performing & that is performed after share reconstruction
            rs_db[i][j] = Bit(bool(rs_bio_share_db[i]->data[j]) ^ bool(raw_query->data[j]), ALICE);
            bp_db[i][j] = Bit(bool(bp_bio_share_db[i]->data[j]), BOB);
            rs_mask_db[i][j] = Bit(rs_bio_share_db[i]->mask[j], ALICE);
            bp_mask_db[i][j] = Bit(bp_bio_share_db[i]->mask[j], BOB);
        }
    }
    for(int j = 0; j < bio.template_size; j++){
        query_mask_bits[j] =  Bit(raw_query->mask[j], ALICE);
    }
    cerr << "  # Setting input finished." << endl;

    // Compute threshold
    Integer distance[bio.db_size];
    Integer query_mask(query_mask_bits);
    for (int i = 0; i < bio.db_size; i++){
        // represent bit vector as long integers to use internal hamming_weight/bitwise functions
        Integer rs = Integer(rs_db[i]);
        Integer bp = Integer(bp_db[i]);
        Integer rs_mask = Integer(rs_mask_db[i]);
        Integer bp_mask = Integer(bp_mask_db[i]);

        auto mask = (rs_mask ^ bp_mask) & query_mask;
        distance[i] = rs ^ bp; // rs == (raw_rs ^ query)
        distance[i] =  distance[i] & mask;

        distance[i] = distance[i].hamming_weight().resize(16);
        auto mask_sz = mask.hamming_weight().resize(16+7);
        mask_sz = mask_sz * T;
        vector<Bit> shifted_mask_bits(mask_sz.bits.begin()+7, mask_sz.bits.end());
        auto frac_thresh = Integer(shifted_mask_bits);
        match[i] = distance[i] < frac_thresh;
    }

    // Fuse matches of f templates for each user and compute the binary membership result
    for (int i = 0; i < bio.user_num; i++){
        Bit fmatch = match[i*bio.fuse];
        for (int j = 0; j < bio.fuse; j++)
            fmatch = fmatch & match[i*bio.fuse + j];
        membership = membership | fmatch;
    }

    cout << "   Membership: " << membership.reveal<bool>(ALICE) << endl;
    cerr << "  # Revealing matches finished." << endl;
    cerr << "db size: " << bio.db_size << ", Biometric size: " << bio.template_size << " total bit size:" << bio.db_size*bio.template_size << endl; 
}


void membership_euc(int party, BioSetting bio) {
    cerr << "Party (" <<  party << ") run Euclidean membership with " << bio.to_str() << endl; 
    // set public parameters
    const int MAX_DIST_LEN = 28;
	Integer T(MAX_DIST_LEN, SIMILARITY_THRESHOLD, PUBLIC);

    // create SMC input variables
	vector<Integer> *rs_db = new vector<Integer>[bio.db_size];
	vector<Integer> *bp_db = new vector<Integer>[bio.db_size];
	vector<Bit> match(bio.db_size);
    Bit membership = Bit(false, ALICE);
    
    // set input values
    for (int i = 0; i < bio.db_size; i++){
        rs_db[i] = vector<Integer>(bio.template_size);
        bp_db[i] = vector<Integer>(bio.template_size);
        for(int j = 0; j < bio.template_size; j++){
            rs_db[i][j] = Integer(bio.bits_per_slot, rs_bio_share_db, ALICE);
            bp_db[i][j] = Integer(bio.bits_per_slot, bp_bio_share_db, BOB);
        }
    }
    cerr << "  # Setting input finished." << endl;

    // Compute threshold
    Integer diff[bio.template_size];
    for (int i = 0; i < bio.db_size; i++){
        Integer distance(MAX_DIST_LEN, 0, ALICE);
        for (int j = 0; j < bio.template_size; j++){
            diff[j] = rs_db[i][j]-bp_db[i][j];
            diff[j].resize(2*bio.bits_per_slot, true);
            diff[j] = diff[j]*diff[j];
            distance = distance + diff[j].resize(MAX_DIST_LEN);
        }
        match[i] = (distance < T);
    }

    for (int i = 0; i < bio.user_num; i++){
        Bit fmatch = match[i*bio.fuse];
        for (int j = 0; j < bio.fuse; j++)
            fmatch = fmatch & match[i*bio.fuse + j];
        membership = membership | fmatch;
    }

    cout << "   Membership: " << membership.reveal<bool>(ALICE) << endl;
    cerr << "  # Revealing matches finished." << endl;
}


int main(int argc, char** argv) {
    string party_str;           // party from  [RS, BP]
    int port;                   // network port
    string addr;                // address of the benchmark file
    string bio_type = "finger"; // biometric type (we optimize comparison for iris to only or MSBs)
    int N = 32;                 // number of users
    int fuse = 1;               // number of templates per user
    int ts = 1;                 // templates size


    CLI::App app{"SMC-Janus"};
    app.add_option("party", party_str, "Party: [RS, BP]")->required();
    app.add_option("port", port, "Network port")->required();
    app.add_option("--bio-type", bio_type, "Biometric type: [finger, iris]");
    app.add_option("--addr", addr, "Address of the output benchmark file");
    app.add_option("-N", N, "Number of registered users (N)");
    app.add_option("-f", fuse, "Number of biometric templates per user (f)");
    app.add_option("--ts", ts, "The size of the biometric template (TS)");

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
        cerr << "Invalid party: " << party_str << endl;
        return 1;
    }  


    ofstream fout;
    fout.open(addr, fout.out | fout.app);

    if (VERBOSE){
        cerr << "party: " << party_str << " (" << party <<  ") port: " << port << " bio-type: " <<  bio_type << " addr: " << addr << " N: " << N << " fuse: " << fuse << endl;
    }


	NetIO * io = new NetIO(party==ALICE ? nullptr : "127.0.0.1", port);
	setup_semi_honest(io, party);

    BioSetting bio(bio_type, N, fuse, ts, SIMILARITY_THRESHOLD, 0, 8);
    gen_data(bio);
    cerr << "Party: " << party_str << " gen_data finished." << endl;

    auto start = chrono::high_resolution_clock::now();
    if (bio.bio_type == "iris"){
        cout << "Running masked hamming membership for iris code." << endl;
        membership_masked_hamming(party, bio); 
    } else{
        cout << "Running Euclidean distance membership." << endl;
        membership_euc(party, bio); 
    }
    auto end = chrono::high_resolution_clock::now();
    auto duration = chrono::duration_cast<chrono::milliseconds>(end-start).count();


    // cout << "$$$$$ Execution finished successfully $$$$$$" << endl;
    auto transfer_byte = io->get_sent_byte_count();
	finalize_semi_honest();
	delete io;


    cout << "Execution of party " << party_str << "finished successfuly.\n" <<
            "    Duration: " << duration << " ms\n" << 
            "    Bytes sent by " << party_str << ": " << transfer_byte << "B. ("  <<
            transfer_byte/1024/1024 << " MB)" <<  endl;

    fout << N << ", " << fuse << ", " << bio.template_size << ", " << duration << ", " << transfer_byte << endl;

    return 0;
}
