# Aim

This golang project is aimed at making life simpler for developing microservices that use mongo as their backend,
providing commonly used fns in a generic yet flexible format. ( It is very similar to java's spring mongorepository )

# Configs

1. Debug tracing enable
2.

# Requirements

1. A generic struct called MongoRepository[T]
2. It should support basic ones that spring supports out of the box like findAll() findById() save() saveAll() deleteById() and any i hve missed here
3. For building custom ones we need to provide a base like findOneCustom() findManyCustom(), deleteOne(), deleteMany(), count() & aggregation() make sure they have correct args for each ( projection , sorting options also ?)
4. Add a new tag option called "index:asc" that will index that field, this will be read from our constructor NewMongoRepository[T] & we will ensureIndexes on construction
